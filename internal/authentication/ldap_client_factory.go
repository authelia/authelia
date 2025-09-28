package authentication

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/go-ldap/ldap/v3"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// LDAPClientFactory an interface describing factories that produce LDAPConnection implementations.
type LDAPClientFactory interface {
	Initialize() (err error)
	GetClient(opts ...LDAPClientFactoryOption) (client ldap.Client, err error)
	ReleaseClient(client ldap.Client) (err error)
	Close() (err error)
}

// NewStandardLDAPClientFactory create a concrete ldap connection factory.
func NewStandardLDAPClientFactory(config *schema.AuthenticationBackendLDAP, certs *x509.CertPool, dialer LDAPClientDialer) LDAPClientFactory {
	if dialer == nil {
		dialer = &LDAPClientDialerStandard{}
	}

	tlsc := utils.NewTLSConfig(config.TLS, certs)

	opts := []ldap.DialOpt{
		ldap.DialWithDialer(&net.Dialer{Timeout: config.Timeout}),
		ldap.DialWithTLSConfig(tlsc),
	}

	return &StandardLDAPClientFactory{
		config: config,
		tls:    tlsc,
		opts:   opts,
		dialer: dialer,
	}
}

// StandardLDAPClientFactory the production implementation of an ldap connection factory.
type StandardLDAPClientFactory struct {
	config *schema.AuthenticationBackendLDAP
	tls    *tls.Config
	opts   []ldap.DialOpt
	dialer LDAPClientDialer
}

func (f *StandardLDAPClientFactory) Initialize() (err error) {
	return nil
}

func (f *StandardLDAPClientFactory) GetClient(opts ...LDAPClientFactoryOption) (client ldap.Client, err error) {
	return getLDAPClient(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts, opts...)
}

func (f *StandardLDAPClientFactory) ReleaseClient(client ldap.Client) (err error) {
	if err = client.Close(); err != nil {
		return fmt.Errorf("error occurred closing LDAP client: %w", err)
	}

	return nil
}

func (f *StandardLDAPClientFactory) Close() (err error) {
	return nil
}

// NewPooledLDAPClientFactory is a decorator for a LDAPClientFactory that performs pooling.
func NewPooledLDAPClientFactory(config *schema.AuthenticationBackendLDAP, certs *x509.CertPool, dialer LDAPClientDialer) (factory LDAPClientFactory) {
	if dialer == nil {
		dialer = &LDAPClientDialerStandard{}
	}

	tlsc := utils.NewTLSConfig(config.TLS, certs)

	opts := []ldap.DialOpt{
		ldap.DialWithDialer(&net.Dialer{Timeout: config.Timeout}),
		ldap.DialWithTLSConfig(tlsc),
	}

	if config.Pooling.Count <= 0 {
		config.Pooling.Count = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Pooling.Count
	}

	if config.Pooling.Retries <= 0 {
		config.Pooling.Retries = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Pooling.Retries
	}

	if config.Pooling.Timeout <= 0 {
		config.Pooling.Timeout = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Pooling.Timeout
	}

	return &PooledLDAPClientFactory{
		config: config,
		tls:    tlsc,
		opts:   opts,
		dialer: dialer,
		minPoolSize:    int32(max(2, config.Pooling.Count/2)),
		maxPoolSize:    int32(max(2, config.Pooling.Count)),
		clientLifetime: time.Minute * 60,
	}
}

// PooledLDAPClientFactory is a LDAPClientFactory that takes another LDAPClientFactory and pools the
// factory generated connections using a channel for thread safety.
type PooledLDAPClientFactory struct {
	config *schema.AuthenticationBackendLDAP
	tls    *tls.Config
	opts   []ldap.DialOpt
	dialer LDAPClientDialer

	pool chan *LDAPClientPooled
	mu   sync.Mutex

	minPoolSize    int32
	maxPoolSize    int32
	clientLifetime time.Duration
	activeCount    int32
	wakeup         chan struct{}
	closed         chan error
	ctx            context.Context
	cancel         context.CancelFunc
	closing bool
}

func (f *PooledLDAPClientFactory) Initialize() (err error) {
	if f.pool != nil {
		return fmt.Errorf("LDAP pool is already initialized")
	}
	f.pool = make(chan *LDAPClientPooled, f.maxPoolSize)
	f.wakeup = make(chan struct{}, 1)
	f.closed = make(chan error, 1)
	f.ctx, f.cancel = context.WithCancel(context.Background())
	if !f.tryAddPooledClient(f.pool) {
		err = fmt.Errorf("LDAP pool couldn't acquire initial client")
	} else if rerr := f.ReadinessCheck(); rerr != nil {
		err = rerr
	} else {
		go f.poolManager(f.ctx, f.pool, f.wakeup, f.closed)
		return nil
	}
	err = fmt.Errorf("LDAP pool initialization failed: %w", err)
	f.closed <- err
	_ = f.Close()
	return err
}

func (f *PooledLDAPClientFactory) ReadinessCheck() (err error) {
	client, err := f.acquire()
	if client != nil {
		return f.ReleaseClient(client)
	}
	return err
}

// GetClient opens new client using the pool.
func (f *PooledLDAPClientFactory) GetClient(opts ...LDAPClientFactoryOption) (conn ldap.Client, err error) {
	if len(opts) != 0 {
		return getLDAPClient(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts, opts...)
	}

	return f.acquire()
}

// The new function creates a pool based client. This function is not thread safe.
func (f *PooledLDAPClientFactory) new() (pooled *LDAPClientPooled, err error) {
	var client ldap.Client

	if client, err = getLDAPClient(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts); err != nil {
		return nil, fmt.Errorf("error occurred establishing new client for the pool: %w", err)
	}

	return &LDAPClientPooled{Client: client}, nil
}

// ReleaseClient returns a client using the pool or closes it.
func (f *PooledLDAPClientFactory) ReleaseClient(client ldap.Client) (err error) {
	c, ok := client.(*LDAPClientPooled)
	if !ok {
		return client.Close()
	}
	if pool := f.pool; pool != nil && c.Client != nil && !c.Client.IsClosing() {
		select {
		case pool <- c:
			return nil
		default:
		}
	}
	return f.disposeClient(c)
}

func (f *PooledLDAPClientFactory) disposeClient(client *LDAPClientPooled) error {
	f.activeCount_Add(-1)
	f.wakeupPoolManager()
	return client.Close()
}

func (f *PooledLDAPClientFactory) acquire() (client *LDAPClientPooled, err error) {
	pool := f.pool
	if pool == nil {
		return nil, fmt.Errorf("error acquiring client: the pool is unitialized or closed")
	}
	deadline := time.Now().Add(f.config.Pooling.Timeout)
	for !f.closing_Load() {
		select {
		case client = <-pool:
		default:
			f.wakeupPoolManager()
			select {
			case client = <-pool:
			case <-time.After(time.Millisecond * 10):
				if time.Now().After(deadline) {
					return nil, fmt.Errorf("error acquiring client: timeout")
				}
				continue
			}
		}
		if len(pool) <= 1 && f.maxPoolSize > 5 {
			f.wakeupPoolManager()
		}
		if client.Client.IsClosing() {
			_ = f.disposeClient(client)
			continue
		}
		return client, nil
	}
	return nil, fmt.Errorf("error acquiring client: the pool is closed")
}

func (f *PooledLDAPClientFactory) tryAddPooledClient(pool chan *LDAPClientPooled) bool {
	attempts := f.config.Pooling.Retries
	sleep := f.config.Timeout / 8
	for !f.closing_Load() {
		if client, _ := f.new(); client != nil {
			select {
			case pool <- client:
				f.activeCount_Add(1)
				return true
			default:
				_ = client.Close()
				return false
			}
		}
		if attempts == 0 {
			return false
		}
		time.Sleep(sleep)
		sleep *= 2
		attempts--
	}
	return false
}

func (f *PooledLDAPClientFactory) wakeupPoolManager() {
	if wakeup := f.wakeup; wakeup != nil {
		select {
		case wakeup <- struct{}{}:
		default:
		}
	}
}

func (f *PooledLDAPClientFactory) poolManager(ctx context.Context, pool chan *LDAPClientPooled, wakeup chan struct{}, result chan error) {
	f.wakeupPoolManager()
	for !f.closing_Load() {
		select {
		case <-ctx.Done():
			result <- ctx.Err()
			return
		case <-wakeup:
			if f.closing_Load() {
				goto poolCleanup
			}
			active := f.activeCount_Load()
			for range min(min(max(f.minPoolSize-active, 2-int32(len(pool))), f.maxPoolSize-active), max(2, (f.maxPoolSize+3)/4)) {
				if !f.tryAddPooledClient(pool) {
					break
				}
			}
		}
	}
poolCleanup:
	for f.activeCount_Load() > 0 {
		select {
		case <-ctx.Done():
			result <- ctx.Err()
			return
		case client, ok := <-pool:
			if !ok {
				result <- fmt.Errorf("channel closed unexpectedly")
				return
			}
			_ = f.disposeClient(client)
		}
	}
	close(result)
}

func (f *PooledLDAPClientFactory) Close() (err error) {
	if f.pool == nil || !f.closing_CompareAndSwap(false, true) {
		return nil
	}
	f.wakeupPoolManager()
	timeout := max(time.Second, 2*f.config.Timeout)
	cleanupCtx, cancelCleanup := context.WithTimeout(context.Background(), timeout)
	defer cancelCleanup()
	select {
	case cleanupErr := <-f.closed:
		err = cleanupErr
	case <-cleanupCtx.Done():
		f.cancel()
		err = fmt.Errorf("pool cleanup exceeded timeout")
	}
	f.pool, f.wakeup, f.closed, f.ctx, f.cancel, cleanupCtx = nil, nil, nil, nil, nil, nil
	return err
}

// atomic.Int32.Load, slow version
func (f *PooledLDAPClientFactory) activeCount_Load() int32 {
	return f.activeCount_Add(0)
}

// atomic.Int32.Add, slow version
func (f *PooledLDAPClientFactory) activeCount_Add(v int32) int32 {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.activeCount += v
	return f.activeCount
}

// atomic.Bool.CompareAndSwap, slow version
func (f *PooledLDAPClientFactory) closing_CompareAndSwap(old, new bool) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closing == old {
		f.closing = new
		return true
	}
	return false
}

// atomic.Bool.Load, slow version
func (f *PooledLDAPClientFactory) closing_Load() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closing
}

// LDAPClientPooled is a decorator for the ldap.Client which handles the pooling functionality. i.e. prevents the client
// from being closed and instead relinquishes the connection back to the pool.
type LDAPClientPooled struct {
	ldap.Client
}

func (c *LDAPClientPooled) Close() error {
	if c.Client != nil {
		err := c.Client.Close()
		c.Client = nil
		return err
	}
	return nil
}

func getLDAPClient(address, username, password string, timeout time.Duration, dialer LDAPClientDialer, tls *tls.Config, startTLS bool, dialerOpts []ldap.DialOpt, opts ...LDAPClientFactoryOption) (client ldap.Client, err error) {
	config := &LDAPClientFactoryOptions{
		Address:  address,
		Username: username,
		Password: password,
	}

	for _, opt := range opts {
		opt(config)
	}

	if client, err = dialer.DialURL(config.Address, dialerOpts...); err != nil {
		return nil, fmt.Errorf("error occurred dialing address: %w", err)
	}

	client.SetTimeout(timeout)

	if tls != nil && startTLS {
		if err = client.StartTLS(tls); err != nil {
			_ = client.Close()

			return nil, fmt.Errorf("error occurred performing starttls: %w", err)
		}
	}

	if config.Password == "" {
		err = client.UnauthenticatedBind(config.Username)
	} else {
		err = client.Bind(config.Username, config.Password)
	}

	if err != nil {
		_ = client.Close()

		return nil, fmt.Errorf("error occurred performing bind: %w", err)
	}

	return client, nil
}
