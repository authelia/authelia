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
	GetClient(opts ...LDAPClientFactoryOption) (client LDAPClient, err error)
	ReleaseClient(client LDAPClient) (err error)
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

func (f *StandardLDAPClientFactory) GetClient(opts ...LDAPClientFactoryOption) (client LDAPClient, err error) {
	return ldapDialBind(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts, opts...)
}

func (f *StandardLDAPClientFactory) ReleaseClient(client LDAPClient) (err error) {
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

	pool chan *PooledLDAPClient

	sleep time.Duration

	mu      sync.Mutex
	next    int
	closing bool
}

func (f *PooledLDAPClientFactory) Initialize() (err error) {
	if f.pool != nil {
		return nil
	}

	f.pool = make(chan *PooledLDAPClient, f.config.Pooling.Count)

	var (
		errs   []error
		client *PooledLDAPClient
	)

	for i := 0; i < f.config.Pooling.Count; i++ {
		if client, err = f.dial(); err != nil {
			errs = append(errs, err)

			f.log.WithError(err).Debug("Error occurred dialing a pooled client")

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
func (f *PooledLDAPClientFactory) GetClient(opts ...LDAPClientFactoryOption) (conn LDAPClient, err error) {
	if len(opts) != 0 {
		f.log.Trace("Dialing new unpooled client")

		return ldapDialBind(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts, opts...)
	}

	return f.acquire(context.Background())
}

// The dial function dials a new LDAPClient and wraps it as a PooledLDAPClient client.
func (f *PooledLDAPClientFactory) dial() (pooled *PooledLDAPClient, err error) {
	var client LDAPClient

	f.log.Trace("Dialing new pooled client")

	if client, err = ldapDialBind(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts); err != nil {
		return nil, fmt.Errorf("error occurred establishing new client for the pool: %w", err)
	}

	f.mu.Lock()

	pooled = &PooledLDAPClient{LDAPClient: client, log: f.log.WithField("client", f.next)}

	f.next++

	f.mu.Unlock()

	pooled.log.Trace("New pooled client created")

	return pooled, nil
}

// ReleaseClient returns a client using the pool or closes it.
func (f *PooledLDAPClientFactory) ReleaseClient(client LDAPClient) (err error) {
	f.log.Trace("Releasing Client")

	if f.isClosing() {
		f.log.Trace("Pooled/Unpooled Client is summarily being closed as the pool is closing or closed")

		return client.Close()
	}

	switch c := client.(type) {
	case *PooledLDAPClient:
		return f.release(c)
	default:
		f.log.Trace("Unpooled client is being closed")

		return client.Close()
	}
}

func (f *PooledLDAPClientFactory) release(client *PooledLDAPClient) (err error) {
	client.log.Trace("Releasing pooled client")

	select {
	case f.pool <- client:
		return nil
	default:
		f.log.Trace("Pooled extra client is being closed")

		return client.Close()
	}
}

func (f *PooledLDAPClientFactory) acquire(ctx context.Context) (client *PooledLDAPClient, err error) {
	f.log.Trace("Acquiring Client")

	if f.isClosing() {
		return nil, NewPoolCtxErr(fmt.Errorf("error acquiring client: the pool is closed"))
	}

	ctx, cancel := context.WithTimeout(ctx, f.config.Pooling.Timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, NewPoolCtxErr(fmt.Errorf("error acquiring client: %w", ctx.Err()))
		case client = <-f.pool:
			if client.healthy() {
				return client, nil
			}

			client.log.Trace("Client is closing or invalid")

			if client, err = f.dial(); err == nil {
				client.log.Trace("New client acquired")

				return client, nil
			}

			f.log.WithError(err).Trace("Error acquiring new client")

			time.Sleep(f.sleep)
		}
	}
}

func (f *PooledLDAPClientFactory) Close() (err error) {
	f.setClosing()

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

func (f *PooledLDAPClientFactory) isClosing() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.closing
}

func (f *PooledLDAPClientFactory) setClosing() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.closing = true
}

// PooledLDAPClient is a decorator for the LDAPClient which handles the pooling functionality. i.e. prevents the client
// from being closed and instead relinquishes the connection back to the pool.
type PooledLDAPClient struct {
	LDAPClient

	log *logrus.Entry
}

func (c *PooledLDAPClient) healthy() bool {
	if c == nil || c.LDAPClient == nil || c.IsClosing() {
		return false
	}

	if _, err := c.WhoAmI(nil); err != nil {
		return false
	}

	return true
}

func ldapDialBind(address, username, password string, timeout time.Duration, dialer LDAPClientDialer, tls *tls.Config, startTLS bool, dialOpts []ldap.DialOpt, opts ...LDAPClientFactoryOption) (client LDAPClient, err error) {
	config := &LDAPClientFactoryOptions{
		Address:  address,
		Username: username,
		Password: password,
	}

	for _, opt := range opts {
		opt(config)
	}

	if client, err = dialer.DialURL(config.Address, dialOpts...); err != nil {
		return nil, fmt.Errorf("error occurred dialing address: %w", err)
	}

	client.SetTimeout(timeout)

	if tls != nil && startTLS {
		if err = client.StartTLS(tls); err != nil {
			_ = client.Close()

			return nil, fmt.Errorf("error occurred performing starttls: %w", err)
		}
	}

	switch {
	case config.Password == "":
		err = client.UnauthenticatedBind(config.Username)
	default:
		err = client.Bind(config.Username, config.Password)
	}

	if err != nil {
		_ = client.Close()

		return nil, fmt.Errorf("error occurred performing bind: %w", err)
	}

	if _, err = client.WhoAmI(nil); err != nil {
		_ = client.Close()

		return nil, fmt.Errorf("error occurred performing whoami: %w", err)
	}

	return client, nil
}
