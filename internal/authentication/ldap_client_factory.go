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
func (f *PooledLDAPClientFactory) GetClient(opts ...LDAPClientFactoryOption) (conn LDAPExtendedClient, err error) {
	if len(opts) != 0 {
		f.log.Trace("Dialing new unpooled client")

		return ldapDialBind(f.log, f.config, f.dialer, f.tls, f.opts, opts...)
	}

	return f.acquire(context.Background())
}

// The dial function dials a new LDAPClient and wraps it as a PooledLDAPClient client.
func (f *PooledLDAPClientFactory) dial() (pooled *PooledLDAPClient, err error) {
	var client LDAPExtendedClient

	f.log.Trace("Dialing new pooled client")

	if client, err = ldapDialBind(f.log, f.config, f.dialer, f.tls, f.opts); err != nil {
		return nil, fmt.Errorf("error occurred establishing new client for the pool: %w", err)
	}

	f.mu.Lock()

	pooled = &PooledLDAPClient{LDAPExtendedClient: client, log: f.log.WithField("client", f.next), permitFeatureDetectionFailure: f.config.PermitFeatureDetectionFailure}

	f.next++

	f.mu.Unlock()

	pooled.log.Trace("New pooled client created")

	return pooled, nil
}

// ReleaseClient returns a client using the pool or closes it.
func (f *PooledLDAPClientFactory) ReleaseClient(client LDAPExtendedClient) (err error) {
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
	LDAPExtendedClient

	log *logrus.Entry

	permitFeatureDetectionFailure bool
}

func (c *PooledLDAPClient) healthy() bool {
	if c == nil || c.LDAPExtendedClient == nil || c.IsClosing() {
		return false
	}

	var (
		err error
	)
	if _, err = ldapGetFeatureSupportFromClient(c); err != nil && !c.permitFeatureDetectionFailure {
		return false
	}

	return true
}

func ldapDialBind(log *logrus.Entry, config *schema.AuthenticationBackendLDAP, dialer LDAPClientDialer, tls *tls.Config, dialOpts []ldap.DialOpt, opts ...LDAPClientFactoryOption) (client LDAPExtendedClient, err error) {
	options := &LDAPClientFactoryOptions{
		Address:  config.Address.String(),
		Username: config.User,
		Password: config.Password,
	}

	for _, opt := range opts {
		opt(options)
	}

	var base LDAPBaseClient
	if base, err = dialer.DialURL(options.Address, dialOpts...); err != nil {
		return nil, fmt.Errorf("error occurred dialing address: %w", err)
	}

	base.SetTimeout(config.Timeout)

	var discovery LDAPDiscovery
	if discovery, err = ldapGetFeatureSupportFromClient(base); err != nil {
		if !config.PermitFeatureDetectionFailure {
			log.WithError(err).Error("Error occurred discovering critical information about server. This may result in reduced functionality or other failures, and is fundamentally not supported.")
		}
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
	//nolint:staticcheck
	switch {
	case options.Password == "":
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
