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
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
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

	sleep := config.Pooling.Timeout / time.Duration(config.Pooling.Retries)

	return &PooledLDAPClientFactory{
		log:    logging.Logger().WithFields(map[string]any{"provider": "pooled ldap factory"}),
		config: config,
		tls:    tlsc,
		opts:   opts,
		dialer: dialer,
		sleep:  sleep,
	}
}

// PooledLDAPClientFactory is a LDAPClientFactory that takes another LDAPClientFactory and pools the
// factory generated connections using a buffered channel for thread safety.
type PooledLDAPClientFactory struct {
	log *logrus.Entry

	config *schema.AuthenticationBackendLDAP
	tls    *tls.Config
	opts   []ldap.DialOpt
	dialer LDAPClientDialer

	pool    chan *LDAPClientPooled
	returns chan *LDAPClientPooled

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu      sync.Mutex
	sleep   time.Duration
	closing bool
}

func (f *PooledLDAPClientFactory) Initialize() (err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.pool != nil {
		return nil
	}

	f.pool = make(chan *LDAPClientPooled, f.config.Pooling.Count)
	f.returns = make(chan *LDAPClientPooled, f.config.Pooling.Count)

	f.ctx, f.cancel = context.WithCancel(context.Background())

	var (
		errs   []error
		client *LDAPClientPooled
	)

	for i := 0; i < f.config.Pooling.Count; i++ {
		if client, err = f.new(); err != nil {
			errs = append(errs, err)

			continue
		}

		f.pool <- client
	}

	if len(errs) == f.config.Pooling.Count {
		return fmt.Errorf("errors occurred initializing the client pool: no connections could be established")
	}

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		f.maintain(f.ctx)
	}()

	return nil
}

// GetClient opens new client using the pool.
func (f *PooledLDAPClientFactory) GetClient(opts ...LDAPClientFactoryOption) (conn ldap.Client, err error) {
	if len(opts) != 0 {
		f.log.Trace("Creating new unpooled client")

		return getLDAPClient(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts, opts...)
	}

	ctx, cancel := context.WithTimeout(context.Background(), f.config.Pooling.Timeout)
	defer cancel()

	select {
	case client := <-f.pool:
		return client, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout waiting for connection from pool: %w", ctx.Err())
	}
}

// The new function creates a pool based client. This function is not thread safe.
func (f *PooledLDAPClientFactory) new() (pooled *LDAPClientPooled, err error) {
	var client ldap.Client

	f.log.Trace("Creating new pooled client")

	if client, err = getLDAPClient(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts); err != nil {
		return nil, fmt.Errorf("error occurred establishing new client for the pool: %w", err)
	}

	pooled = &LDAPClientPooled{Client: client}

	return pooled, nil
}

// ReleaseClient returns a client using the pool or closes it.
func (f *PooledLDAPClientFactory) ReleaseClient(client ldap.Client) (err error) {
	pooled, ok := client.(*LDAPClientPooled)
	if !ok {
		return client.Close()
	}

	select {
	case f.returns <- pooled:
		return nil
	default:
		return client.Close()
	}
}

func (f *PooledLDAPClientFactory) maintain(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case client := <-f.returns:
			if client.IsClosing() || client.Client == nil {
				client.Close()
				if newClient, err := f.new(); err != nil {
					select {
					case f.pool <- client:
					default:
						newClient.Close()
					}
				}
			} else {
				select {
				case f.pool <- client:
				default:
					client.Close()
				}
			}
		case <-ticker.C:
			needed := cap(f.pool) - len(f.pool)
			for i := 0; i < needed; i++ {
				if client, err := f.new(); err != nil {
					select {
					case f.pool <- client:
					default:
						client.Close()
						return
					}
				} else {
					break
				}
			}
		}
	}
}

func (f *PooledLDAPClientFactory) Close() (err error) {
	f.mu.Lock()
	f.closing = true
	f.mu.Unlock()

	if f.cancel != nil {
		f.cancel()
	}

	f.wg.Wait()

	var errs []error

	// Drain and close pool
	if f.pool != nil {
		close(f.pool)
		for client := range f.pool {
			if !client.IsClosing() {
				if err = client.Close(); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}

	// Drain and close returns
	if f.returns != nil {
		close(f.returns)
		for client := range f.returns {
			if !client.IsClosing() {
				if err = client.Close(); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf("errors occurred closing the client pool: %w", errs[0])
	}

	return nil
}

// LDAPClientPooled is a decorator for the ldap.Client which handles the pooling functionality. i.e. prevents the client
// from being closed and instead relinquishes the connection back to the pool.
type LDAPClientPooled struct {
	ldap.Client
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
