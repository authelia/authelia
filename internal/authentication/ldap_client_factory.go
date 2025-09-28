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

	factory = &PooledLDAPClientFactory{
		config:         config,
		tls:            tlsc,
		opts:           opts,
		dialer:         dialer,
		minPoolSize:    int32(max(2, config.Pooling.Count/2)),
		maxPoolSize:    int32(max(2, config.Pooling.Count)),
		clientLifetime: time.Minute * 60,
	}

	return factory
}

// PooledLDAPClientFactory is a LDAPClientFactory that takes another LDAPClientFactory and pools the
// factory generated connections using a channel for thread safety.
type PooledLDAPClientFactory struct {
	config         *schema.AuthenticationBackendLDAP
	tls            *tls.Config
	opts           []ldap.DialOpt
	dialer         LDAPClientDialer
	minPoolSize    int32
	maxPoolSize    int32
	clientLifetime time.Duration
	pool           chan *LDAPClientPooled
	activeCount    int32
	wakeup         chan struct{}
	closing        bool
	closed         chan error
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.Mutex
}

// atomic.Int32.Load, slow version
func (f *PooledLDAPClientFactory) activeCountLoad() int32 {
	return f.activeCountAdd(0)
}

// atomic.Int32.Add, slow version
func (f *PooledLDAPClientFactory) activeCountAdd(v int32) int32 {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.activeCount += v
	return f.activeCount
}

// atomic.Bool.CompareAndSwap, slow version
func (f *PooledLDAPClientFactory) closingCompareAndSwap(old, new bool) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closing == old {
		f.closing = new
		return true
	}
	return false
}

// atomic.Bool.Load, slow version
func (f *PooledLDAPClientFactory) closingLoad() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closing
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
		go f.poolManager()
		return nil
	}
	err = fmt.Errorf("LDAP pool initialization failed: %w", err)
	f.closed <- err
	_ = f.Close()
	return err
}

func (f *PooledLDAPClientFactory) ReadinessCheck() (err error) {
	client, err := f.acquire()
	if err != nil {
		return err
	}
	_ = f.ReleaseClient(client)
	return nil
}

func (f *PooledLDAPClientFactory) GetClient(opts ...LDAPClientFactoryOption) (conn ldap.Client, err error) {
	if len(opts) != 0 {
		return getLDAPClient(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts, opts...)
	}
	return f.acquire()
}

// The new function creates a pool based client. This function is not thread safe.
func (f *PooledLDAPClientFactory) new() (pooled *LDAPClientPooled, err error) {
	client, err := getLDAPClient(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts)
	if err != nil {
		return nil, err
	}

	return &LDAPClientPooled{Client: client}, nil
}

func (f *PooledLDAPClientFactory) ReleaseClient(client ldap.Client) (err error) {
	p, ok := client.(*LDAPClientPooled)
	if !ok {
		return client.Close()
	}
	f_pool := f.pool
	if f_pool == nil || !p.IsHealthy() {
		return f.disposeClient(p)
	}
	select {
	case f_pool <- p:
		return nil
	default:
		return f.disposeClient(p)
	}
}

func (f *PooledLDAPClientFactory) disposeClient(client *LDAPClientPooled) error {
	if client == nil {
		return nil
	}
	f.activeCountAdd(-1)
	f.wakeupPoolManager()
	return client.Close()
}

func (f *PooledLDAPClientFactory) acquire() (client *LDAPClientPooled, err error) {
	f_pool := f.pool
	if f_pool == nil {
		return nil, fmt.Errorf("pool is unitialized or closed")
	}
	deadline := time.Now().Add(f.config.Pooling.Timeout)
	for !f.closingLoad() {
		select {
		case client = <-f_pool:
			goto check_client
		default:
			f.wakeupPoolManager()
			select {
			case client = <-f_pool:
				goto check_client
			case <-time.After(time.Millisecond * 10):
				if time.Now().After(deadline) {
					return nil, fmt.Errorf("timeout waiting for client from LDAP pool")
				}
				continue
			}
		}
	check_client:
		if len(f_pool) <= 1 && f.maxPoolSize > 5 {
			f.wakeupPoolManager()
		}
		if !client.IsHealthy() {
			_ = f.disposeClient(client)
			continue
		}
		return client, nil
	}
	return nil, fmt.Errorf("LDAP pool is closing, cannot acquire client")
}

func (f *PooledLDAPClientFactory) tryAddPooledClient(f_pool chan *LDAPClientPooled) bool {
	attempts := f.config.Pooling.Retries
	sleep := f.config.Timeout / 8
	for !f.closingLoad() {
		client, err := f.new()
		if err == nil {
			select {
			case f_pool <- client:
				f.activeCountAdd(1)
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
	f_wakeup := f.wakeup
	if f_wakeup == nil {
		return
	}
	select {
	case f_wakeup <- struct{}{}:
	default:
	}
}

func (f *PooledLDAPClientFactory) poolManager() {
	f_ctx, f_pool, f_wakeup, f_closed := f.ctx, f.pool, f.wakeup, f.closed
	f.wakeupPoolManager()
	for !f.closingLoad() {
		select {
		case <-f_ctx.Done():
			f_closed <- fmt.Errorf("pool context cancelled")
			return
		case <-f_wakeup:
			if f.closingLoad() {
				goto poolCleanup
			}
			active := f.activeCountLoad()
			available := int32(len(f_pool))
			request := min(max(f.minPoolSize-active, 2-available), f.maxPoolSize-active)
			if request <= 0 {
				continue
			}
			request = min(request, max(2, (f.minPoolSize+1)/2))
			for range request {
				if !f.tryAddPooledClient(f_pool) {
					break
				}
			}
		}
	}
poolCleanup:
	for f.activeCountLoad() > 0 {
		select {
		case <-f_ctx.Done():
			f_closed <- fmt.Errorf("pool context cancelled")
			return
		case client, ok := <-f_pool:
			if !ok {
				f_closed <- fmt.Errorf("pool channel unexpectedly closed")
				return
			}
			_ = f.disposeClient(client)
		}
	}
	close(f_closed)
}

func (f *PooledLDAPClientFactory) Close() (err error) {
	if f.pool == nil || !f.closingCompareAndSwap(false, true) {
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
	f.pool = nil
	f.wakeup = nil
	f.closed = nil
	f.ctx = nil
	f.cancel = nil
	cleanupCtx = nil
	return err
}

type LDAPClientPooled struct {
	ldap.Client
}

func (c *LDAPClientPooled) IsHealthy() bool {
	return c.Client != nil && !c.Client.IsClosing()
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
