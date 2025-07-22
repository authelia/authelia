package authentication

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
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

	pool chan *LDAPClientPooled
	mu   sync.Mutex

	sleep time.Duration

	next int

	closing bool
}

func (f *PooledLDAPClientFactory) Initialize() (err error) {
	f.mu.Lock()

	defer f.mu.Unlock()

	if f.pool != nil {
		return nil
	}

	f.pool = make(chan *LDAPClientPooled, f.config.Pooling.Count)

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

	return nil
}

// GetClient opens new client using the pool.
func (f *PooledLDAPClientFactory) GetClient(opts ...LDAPClientFactoryOption) (conn ldap.Client, err error) {
	if len(opts) != 0 {
		f.log.Trace("Creating new unpooled client")

		return getLDAPClient(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts, opts...)
	}

	return f.acquire(context.Background())
}

// The new function creates a pool based client. This function is not thread safe.
func (f *PooledLDAPClientFactory) new() (pooled *LDAPClientPooled, err error) {
	var client ldap.Client

	f.log.Trace("Creating new pooled client")

	if client, err = getLDAPClient(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts); err != nil {
		return nil, fmt.Errorf("error occurred establishing new client for the pool: %w", err)
	}

	pooled = &LDAPClientPooled{Client: client, log: f.log.WithField("client", f.next)}

	f.next++

	pooled.log.Trace("New pooled client created")

	return pooled, nil
}

// ReleaseClient returns a client using the pool or closes it.
func (f *PooledLDAPClientFactory) ReleaseClient(client ldap.Client) (err error) {
	f.log.Trace("Releasing Client")

	f.mu.Lock()

	defer f.mu.Unlock()

	if f.closing {
		f.log.Trace("Pool is closing, closing the released client")

		return client.Close()
	}

	var (
		pool *LDAPClientPooled
		ok   bool
	)
	if pool, ok = client.(*LDAPClientPooled); !ok {
		f.log.Trace("Unpooled client is being closed")

		// Prevent extra or non-pool connections from being returned into the pool.
		return client.Close()
	}

	pool.log.Trace("Releasing pooled client")

	select {
	case f.pool <- pool:
		pool.log.Trace("Returning pooled client to the pool")

		return nil
	default:
		pool.log.Trace("Closing extra pooled client")

		return client.Close()
	}
}

func (f *PooledLDAPClientFactory) acquire(ctx context.Context) (client *LDAPClientPooled, err error) {
	f.log.Trace("Acquiring Client")

	f.mu.Lock()

	defer f.mu.Unlock()

	if f.closing {
		return nil, fmt.Errorf("error acquiring client: the pool is closed")
	}

	if cap(f.pool) != f.config.Pooling.Count {
		if err = f.Initialize(); err != nil {
			return nil, fmt.Errorf("error acquiring client: error initializing buffer: %w", err)
		}
	}

	f.log.Trace("Timeout Started")

	ctx, cancel := context.WithTimeout(ctx, f.config.Pooling.Timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("error acquiring client: %w", err)
		case client = <-f.pool:
			if client.IsClosing() || client.Client == nil {
				client.log.Trace("Client is closing or invalid")

				if client, err = f.new(); err != nil {
					f.log.WithError(err).Trace("Error acquiring new client")

					time.Sleep(f.sleep)

					continue
				}

				client.log.Trace("New client acquired")
			}

			return client, nil
		}
	}
}

func (f *PooledLDAPClientFactory) Close() (err error) {
	f.mu.Lock()

	defer f.mu.Unlock()

	f.closing = true

	close(f.pool)

	var errs []error

	for client := range f.pool {
		if client.IsClosing() {
			continue
		}

		if err = client.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred closing the client pool: %w", errors.Join(errs...))
	}

	return nil
}

// LDAPClientPooled is a decorator for the ldap.Client which handles the pooling functionality. i.e. prevents the client
// from being closed and instead relinquishes the connection back to the pool.
type LDAPClientPooled struct {
	ldap.Client

	log *logrus.Entry
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
