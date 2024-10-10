package authentication

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"github.com/go-ldap/ldap/v3"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// LDAPClientFactory an interface describing factories that produce LDAPConnection implementations.
type LDAPClientFactory interface {
	Initialize() (err error)
	GetClient(opts ...LDAPClientFactoryOption) (client ldap.Client, err error)
	Shutdown() (err error)
}

// NewLDAPClientFactoryStandard create a concrete ldap connection factory.
func NewLDAPClientFactoryStandard(config *schema.AuthenticationBackendLDAP, certs *x509.CertPool, dialer LDAPClientDialer) *LDAPClientFactoryStandard {
	if dialer == nil {
		dialer = &LDAPClientDialerStandard{}
	}

	tlsc := utils.NewTLSConfig(config.TLS, certs)

	opts := []ldap.DialOpt{
		ldap.DialWithDialer(&net.Dialer{Timeout: config.Timeout}),
		ldap.DialWithTLSConfig(tlsc),
	}

	return &LDAPClientFactoryStandard{
		config: config,
		tls:    tlsc,
		opts:   opts,
		dialer: dialer,
	}
}

// LDAPClientFactoryStandard the production implementation of an ldap connection factory.
type LDAPClientFactoryStandard struct {
	config *schema.AuthenticationBackendLDAP
	tls    *tls.Config
	opts   []ldap.DialOpt
	dialer LDAPClientDialer
}

func (f *LDAPClientFactoryStandard) Initialize() (err error) {
	return nil
}

func (f *LDAPClientFactoryStandard) GetClient(opts ...LDAPClientFactoryOption) (client ldap.Client, err error) {
	config := &LDAPClientFactoryOptions{
		Address:  f.config.Address.String(),
		Username: f.config.User,
		Password: f.config.Password,
	}

	for _, opt := range opts {
		opt(config)
	}

	if client, err = f.dialer.DialURL(config.Address, f.opts...); err != nil {
		return nil, fmt.Errorf("error occurred dialing address: %w", err)
	}

	if f.tls != nil && f.config.StartTLS {
		if err = client.StartTLS(f.tls); err != nil {
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

func (f *LDAPClientFactoryStandard) Shutdown() (err error) {
	return nil
}

// NewLDAPConnectionFactoryPooled is a decorator for a LDAPClientFactory that performs pooling.
func NewLDAPConnectionFactoryPooled(factory LDAPClientFactory, count, retries int, timeout time.Duration) (pool *LDAPClientFactoryPooled) {
	if count <= 0 {
		count = 5
	}

	if retries <= 0 {
		retries = 2
	}

	if timeout.Seconds() <= 0 {
		timeout = time.Second * 10
	}

	sleep := timeout / time.Duration(retries)

	return &LDAPClientFactoryPooled{
		factory: factory,
		count:   count,
		timeout: timeout,
		sleep:   sleep,
	}
}

// LDAPClientFactoryPooled is a LDAPClientFactory that takes another LDAPClientFactory and pools the
// factory generated connections using a channel for thread safety.
type LDAPClientFactoryPooled struct {
	factory LDAPClientFactory

	count  int
	active int

	timeout time.Duration
	sleep   time.Duration

	clients chan *LDAPClientPooled

	closing bool
}

func (f *LDAPClientFactoryPooled) Initialize() (err error) {
	f.clients = make(chan *LDAPClientPooled, f.count)

	var (
		errs   []error
		client *LDAPClientPooled
	)

	for i := 0; i < f.count; i++ {
		if client, err = f.new(); err != nil {
			errs = append(errs, err)

			continue
		}

		f.clients <- client
	}

	if len(errs) == f.count {
		return fmt.Errorf("errors occurred initializing the client pool: no connections could be established")
	}

	return nil
}

// GetClient opens new client using the pool.
func (f *LDAPClientFactoryPooled) GetClient(opts ...LDAPClientFactoryOption) (conn ldap.Client, err error) {
	if len(opts) != 0 {
		return f.factory.GetClient(opts...)
	}

	return f.acquire(context.Background())
}

func (f *LDAPClientFactoryPooled) new() (pooled *LDAPClientPooled, err error) {
	c, l := cap(f.clients), len(f.clients)

	if f.active >= f.count || (c != 0 && c == l) {
		return nil, fmt.Errorf("error occurred establishing new client for the pool: pool is already the maximum size")
	}

	var client ldap.Client

	if client, err = f.factory.GetClient(); err != nil {
		return nil, err
	}

	f.active += 1

	return &LDAPClientPooled{pool: f, Client: client}, nil
}

func (f *LDAPClientFactoryPooled) relinquish(client *LDAPClientPooled) (err error) {
	if f.closing {
		return client.Client.Close()
	}

	// Prevent extra connections from being added to the f and hanging around.
	if cap(f.clients) == len(f.clients) {
		return client.Client.Close()
	}

	f.clients <- client

	return nil
}

func (f *LDAPClientFactoryPooled) acquire(ctx context.Context) (client *LDAPClientPooled, err error) {
	if f.closing {
		return nil, fmt.Errorf("error acquiring client: the pool is closed")
	}

	if cap(f.clients) != f.count {
		if err = f.Initialize(); err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case client = <-f.clients:
		if client.IsClosing() || client.Client == nil {
			f.active -= 1

			for {
				if client, err = f.new(); err != nil {
					time.Sleep(f.sleep)

					continue
				}

				return client, nil
			}
		}

		return client, nil
	}
}

func (f *LDAPClientFactoryPooled) Shutdown() (err error) {
	f.closing = true

	close(f.clients)

	for client := range f.clients {
		_ = client.Client.Close()
	}

	return nil
}

// LDAPClientPooled is a decorator for the ldap.Client which handles the pooling functionality. i.e. prevents the client
// from being closed and instead relinquishes the connection back to the pool.
type LDAPClientPooled struct {
	pool *LDAPClientFactoryPooled

	ldap.Client
}

// Close the LDAPClientPooled by relinquishing access to it and making it available in the pool again.
func (c *LDAPClientPooled) Close() (err error) {
	client, pool := c.Client, c.pool

	// We dereference here to prevent this struct from being reused in an improper way.
	// Messing up the connection state or using it in two routines would be much worse than a panic which we can
	// recover() from.
	c.pool, c.Client = nil, nil

	return pool.relinquish(&LDAPClientPooled{pool: pool, Client: client})
}
