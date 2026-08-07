package authentication

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
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
	GetClient(opts ...LDAPClientFactoryOption) (client LDAPExtendedClient, err error)
	ReleaseClient(client LDAPExtendedClient) (err error)
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
		log:    logging.Logger().WithFields(map[string]any{"provider": "standard ldap factory"}),
		config: config,
		tls:    tlsc,
		opts:   opts,
		dialer: dialer,
	}
}

// StandardLDAPClientFactory the production implementation of an ldap connection factory.
type StandardLDAPClientFactory struct {
	log    *logrus.Entry
	config *schema.AuthenticationBackendLDAP
	tls    *tls.Config
	opts   []ldap.DialOpt
	dialer LDAPClientDialer
}

func (f *StandardLDAPClientFactory) Initialize() (err error) {
	return nil
}

func (f *StandardLDAPClientFactory) GetClient(opts ...LDAPClientFactoryOption) (client LDAPExtendedClient, err error) {
	return ldapDialBind(f.log, f.config, f.dialer, f.tls, f.opts, opts...)
}

func (f *StandardLDAPClientFactory) ReleaseClient(client LDAPExtendedClient) (err error) {
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

	count, retries, timeout := config.Pooling.Count, config.Pooling.Retries, config.Pooling.Timeout

	if count <= 0 {
		count = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Pooling.Count
	}

	if retries <= 0 {
		retries = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Pooling.Retries
	}

	if timeout <= 0 {
		timeout = schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Pooling.Timeout
	}

	return &PooledLDAPClientFactory{
		log:     logging.Logger().WithFields(map[string]any{"provider": "pooled ldap factory"}),
		config:  config,
		tls:     tlsc,
		opts:    opts,
		dialer:  dialer,
		count:   count,
		timeout: timeout,
		sleep:   timeout / time.Duration(retries),
		done:    make(chan struct{}),
		wake:    make(chan struct{}, 1),
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
	done chan struct{}
	wake chan struct{}

	mu      sync.Mutex
	closing bool

	count int
	size  int
	next  atomic.Uint64

	timeout time.Duration
	sleep   time.Duration
}

// Initialize fills the pool with clients. It's safe to call multiple times, only the first call has an effect.
func (f *PooledLDAPClientFactory) Initialize() (err error) {
	f.mu.Lock()

	if f.closing {
		f.mu.Unlock()

		return ErrPoolClosedInitialize
	}

	if f.pool != nil {
		f.mu.Unlock()

		return nil
	}

	f.pool = make(chan *PooledLDAPClient, f.count)

	f.mu.Unlock()

	var (
		errs   []error
		client *PooledLDAPClient
	)

	for range f.count {
		if client, err = f.dial(); err != nil {
			errs = append(errs, err)

			f.log.WithError(err).Debug("Error occurred dialing a pooled client")

			continue
		}

		f.mu.Lock()

		if f.closing {
			f.mu.Unlock()

			if err = client.Close(); err != nil {
				f.log.WithError(err).Debug("Error occurred closing a pooled client dialed while the pool was closing")
			}

			return ErrPoolClosedInitialize
		}

		f.pool <- client

		f.size++

		f.mu.Unlock()
	}

	if len(errs) == f.count {
		return fmt.Errorf("errors occurred initializing the client pool: no connections could be established")
	}

	return nil
}

// GetClient opens new client using the pool.
func (f *PooledLDAPClientFactory) GetClient(opts ...LDAPClientFactoryOption) (conn LDAPExtendedClient, err error) {
	if unpooled, options := f.poolOptions(opts...); unpooled {
		f.log.Trace("Dialing new unpooled client")

		return ldapDialBindOpts(f.log, f.config, f.dialer, f.tls, f.opts, options)
	}

	return f.acquire(context.Background())
}

func (f *PooledLDAPClientFactory) poolOptions(opts ...LDAPClientFactoryOption) (unpooled bool, options *LDAPClientFactoryOptions) {
	if len(opts) == 0 {
		return false, &LDAPClientFactoryOptions{}
	}

	options = ldapClientFactoryOptions(f.config, opts...)

	unpooled = options.Address != f.config.Address.String() || options.Username != f.config.User || options.Password != f.config.Password

	return unpooled, options
}

func (f *PooledLDAPClientFactory) dial() (pooled *PooledLDAPClient, err error) {
	var client LDAPExtendedClient

	f.log.Trace("Dialing new pooled client")

	if client, err = ldapDialBind(f.log, f.config, f.dialer, f.tls, f.opts, WithPermitUnauthenticatedBind(f.config.PermitUnauthenticatedBind)); err != nil {
		return nil, fmt.Errorf("error occurred establishing new client for the pool: %w", err)
	}

	pooled = &PooledLDAPClient{LDAPExtendedClient: client, log: f.log.WithField("client", f.next.Add(1)-1)}

	pooled.log.Trace("New pooled client created")

	return pooled, nil
}

// ReleaseClient returns a client using the pool or closes it.
func (f *PooledLDAPClientFactory) ReleaseClient(client LDAPExtendedClient) (err error) {
	f.log.Trace("Releasing Client")

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

	f.mu.Lock()

	if f.closing {
		client.log.Trace("Pooled client is being closed as the pool is closing or closed")
	} else {
		select {
		case f.pool <- client:
			f.mu.Unlock()

			return nil
		default:
			f.log.Trace("Pooled extra client is being closed")
		}
	}

	f.mu.Unlock()

	f.forget()

	return client.Close()
}

func (f *PooledLDAPClientFactory) forget() {
	f.mu.Lock()

	if f.size > 0 {
		f.size--
	}

	f.mu.Unlock()

	select {
	case f.wake <- struct{}{}:
	default:
	}
}

func (f *PooledLDAPClientFactory) grow() (client *PooledLDAPClient, err error) {
	f.mu.Lock()

	if f.closing || f.size >= f.count {
		f.mu.Unlock()

		return nil, nil
	}

	f.size++

	f.mu.Unlock()

	if client, err = f.dial(); err != nil {
		f.forget()

		return nil, err
	}

	f.mu.Lock()

	if f.closing {
		f.mu.Unlock()

		f.forget()

		if err = client.Close(); err != nil {
			client.log.WithError(err).Trace("Error occurred closing a client dialed while the pool was closing")
		}

		return nil, nil
	}

	f.mu.Unlock()

	return client, nil
}

func (f *PooledLDAPClientFactory) refresh(client *PooledLDAPClient) (replacement *PooledLDAPClient) {
	var err error

	client.log.Trace("Client is closing or invalid")

	if err = client.Close(); err != nil {
		client.log.WithError(err).Trace("Error occurred closing unhealthy client")
	}

	if replacement, err = f.dial(); err == nil {
		replacement.log.Trace("New client acquired")

		return replacement
	}

	f.log.WithError(err).Trace("Error acquiring new client")

	if err = f.release(client); err != nil {
		client.log.WithError(err).Trace("Error occurred returning unhealthy client to the pool")
	}

	return nil
}

func (f *PooledLDAPClientFactory) backoff(ctx context.Context) (err error) {
	select {
	case <-ctx.Done():
		return NewPoolCtxErr(fmt.Errorf("error acquiring client: %w", ctx.Err()))
	case <-f.done:
		return ErrPoolClosed
	case <-time.After(f.sleep):
		return nil
	}
}

func (f *PooledLDAPClientFactory) acquire(ctx context.Context) (client *PooledLDAPClient, err error) {
	f.log.Trace("Acquiring Client")

	if f.isClosing() {
		return nil, ErrPoolClosed
	}

	if len(f.pool) == 0 {
		if client, err = f.grow(); err != nil {
			f.log.WithError(err).Trace("Error occurred growing the pool")
		} else if client != nil {
			client.log.Trace("New client acquired by growing the pool")

			return client, nil
		}
	}

	ctx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, NewPoolCtxErr(fmt.Errorf("error acquiring client: %w", ctx.Err()))
		case <-f.done:
			return nil, ErrPoolClosed
		case client = <-f.pool:
			if client.healthy() {
				return client, nil
			}

			if replacement := f.refresh(client); replacement != nil {
				return replacement, nil
			}

			if err = f.backoff(ctx); err != nil {
				return nil, err
			}
		}
	}
}

func (f *PooledLDAPClientFactory) drain(pool chan *PooledLDAPClient) (errs []error) {
	var err error

	for {
		select {
		case client := <-pool:
			f.forget()

			if client.IsClosing() {
				continue
			}

			if err = client.Close(); err != nil {
				errs = append(errs, err)
			}
		default:
			return errs
		}
	}
}

func (f *PooledLDAPClientFactory) wait() (err error) {
	timer := time.NewTimer(f.timeout)

	defer timer.Stop()

	for {
		f.mu.Lock()

		outstanding := f.size

		f.mu.Unlock()

		if outstanding == 0 {
			return nil
		}

		select {
		case <-f.wake:
		case <-timer.C:
			return fmt.Errorf("timeout of %s elapsed waiting for %d checked out clients to be released", f.timeout, outstanding)
		}
	}
}

// Close drains the pool and closes every client, then waits for any checked out clients, up to the pooling timeout.
func (f *PooledLDAPClientFactory) Close() (err error) {
	f.mu.Lock()

	if f.closing {
		f.mu.Unlock()

		return nil
	}

	f.closing = true

	close(f.done)

	pool := f.pool

	f.mu.Unlock()

	errs := f.drain(pool)

	if err = f.wait(); err != nil {
		errs = append(errs, err)
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

// PooledLDAPClient is a decorator for the LDAPClient which associates a client with the pool that owns it.
type PooledLDAPClient struct {
	LDAPExtendedClient

	log *logrus.Entry
}

func (c *PooledLDAPClient) healthy() bool {
	if c == nil || c.LDAPExtendedClient == nil || c.IsClosing() {
		return false
	}

	var (
		discovery LDAPDiscovery
		err       error
	)

	if discovery, err = ldapGetFeatureSupportFromClient(c); err != nil {
		return false
	}

	if setter, ok := c.LDAPExtendedClient.(ldapDiscoverySetter); ok {
		setter.setDiscovery(discovery)
	}

	return true
}

func ldapClientFactoryOptions(config *schema.AuthenticationBackendLDAP, opts ...LDAPClientFactoryOption) (options *LDAPClientFactoryOptions) {
	options = &LDAPClientFactoryOptions{
		Address:  config.Address.String(),
		Username: config.User,
		Password: config.Password,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

func ldapDialBind(log *logrus.Entry, config *schema.AuthenticationBackendLDAP, dialer LDAPClientDialer, tls *tls.Config, dialOpts []ldap.DialOpt, opts ...LDAPClientFactoryOption) (client LDAPExtendedClient, err error) {
	options := ldapClientFactoryOptions(config, opts...)

	return ldapDialBindOpts(log, config, dialer, tls, dialOpts, options)
}

func ldapDialBindOpts(log *logrus.Entry, config *schema.AuthenticationBackendLDAP, dialer LDAPClientDialer, tls *tls.Config, dialOpts []ldap.DialOpt, options *LDAPClientFactoryOptions) (client LDAPExtendedClient, err error) {
	var base LDAPBaseClient
	if base, err = dialer.DialURL(options.Address, dialOpts...); err != nil {
		return nil, fmt.Errorf("error occurred dialing address: %w", err)
	}

	base.SetTimeout(config.Timeout)

	var discovery LDAPDiscovery
	if discovery, err = ldapGetFeatureSupportFromClient(base); err != nil {
		log.WithError(err).Trace("Error occurred discovering critical information about server. This may result in reduced functionality or other failures, and is fundamentally not supported.")
	}

	client = &LDAPClient{
		LDAPBaseClient: base,
		discovery:      discovery,
	}

	if tls != nil && !config.Address.IsExplicitlySecure() && config.StartTLS {
		if err = client.StartTLS(tls); err != nil {
			_ = client.Close()

			return nil, fmt.Errorf("error occurred performing starttls: %w", err)
		}
	}

	// TODO: Add additional bind logic here, such as MD5Bind, NTLMBind, NTLMUnauthenticatedBind, etc.
	switch {
	case options.Password == "" && options.PermitUnauthenticatedBind:
		err = client.UnauthenticatedBind(options.Username)
	default:
		err = client.Bind(options.Username, options.Password)
	}

	if err != nil {
		_ = client.Close()

		return nil, fmt.Errorf("error occurred performing bind: %w", err)
	}

	return client, nil
}
