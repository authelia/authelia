package authentication

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/rand"
	"net"
	"sync/atomic"
	"time"

	"github.com/go-ldap/ldap/v3"

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

	factory = &PooledLDAPClientFactory{
		config:        	config,
		tls:           	tlsc,
		opts:          	opts,
		dialer:        	dialer,
		// these could possibly be configured, or initialized differently
		minPoolSize:    int32(max(2, config.Pooling.Count/2)),
		maxPoolSize:    int32(max(2, config.Pooling.Count)),
		clientLifetime: time.Hour, // this is a soft target, actual lifetime is fuzzed slightly
	}

	return factory
}

// PooledLDAPClientFactory is a LDAPClientFactory that takes another LDAPClientFactory and pools the
// factory generated connections using a channel for thread safety.
type PooledLDAPClientFactory struct {
	config *schema.AuthenticationBackendLDAP
	tls    *tls.Config
	opts   []ldap.DialOpt
	dialer LDAPClientDialer

	// Pool management
	pool           chan *LDAPClientPooled // Channel for available clients
	activeCount    atomic.Int32           // Atomic counter for active connections
	minPoolSize    int32                  // Minimum number of pool connections to maintain
	maxPoolSize    int32                  // Maximum number of pool connections
	clientLifetime time.Duration          // Target lifetime for a pooled client

	// Synchronization
	wakeup  chan struct{}      // Channel for client requests
	closing atomic.Bool        // Atomic flag indicating if the pool is closing
	closed  chan error         // Signals when poolManager has completed cleanup
	ctx     context.Context    // Context for cancellation
	cancel  context.CancelFunc // Function to cancel the context
}

func (f *PooledLDAPClientFactory) isClosing() bool {
	return f.closing.Load()
}

func (f *PooledLDAPClientFactory) Initialize() (err error) {
	if f.pool != nil { // Already initialized
		return fmt.Errorf("LDAP pool is already initialized")
	}

	// TODO: validate config.Timeout, config.Pooling.Timeout, config.Pooling.Retries
	//       unless "fail fast" behavior is desired, config.Pooling.Timeout should be
	// 			 sufficiently larger than ( config.Timeout * config.Pooling.Retries ) to
	//			 allow worst-case retry scenarios.

	f.pool = make(chan *LDAPClientPooled, f.maxPoolSize)
	f.wakeup = make(chan struct{}, 1)
	f.closed = make(chan error, 1)

	f.ctx, f.cancel = context.WithCancel(context.Background())

	// prime the pool to test connectivity and run a quick self-check
	if !f.tryAddPooledClient() {
		err = fmt.Errorf("LDAP pool couldn't acquire initial client")
	} else if rerr := f.ReadinessCheck(); rerr != nil {
		err = rerr // failed readiness check
	} else {
		go f.poolManager() // transfer responsibility to the pool manager
		return nil
	}

	logging.Logger().Debug("LDAP pool initialization failed, aborting...")
	close(f.closed)
	_ = f.Close()
	return fmt.Errorf("LDAP pool initialization failed: %w", err)
}

// GetClient opens new client using the pool.
func (f *PooledLDAPClientFactory) GetClient(opts ...LDAPClientFactoryOption) (conn ldap.Client, err error) {
	if len(opts) != 0 {
		return getLDAPClient(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts, opts...)
	}

	return f.acquire()
}

// The new function creates a pool based client. This function is not thread safe.
// NOTE: this function adheres to config.Timeout, not config.Pooling.Timeout!
func (f *PooledLDAPClientFactory) new() (pooled *LDAPClientPooled, err error) {
	client, err := getLDAPClient(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts)
	if err != nil { return nil, err }

	return &LDAPClientPooled{Client: client}, nil
}

// ReleaseClient returns a client using the pool or closes it.
func (f *PooledLDAPClientFactory) ReleaseClient(client ldap.Client) (err error) {
	logging.Logger().Trace("Releasing LDAP client")

	pooled, ok := client.(*LDAPClientPooled)
	if !ok {
		logging.Logger().Debug("LDAP client is not pooled, closing directly")
		return client.Close()
	}

	if !pooled.IsHealthy() || pooled.IsExpired() {
		logging.Logger().Info("Pooled LDAP client is not healthy or has expired, disposing")
		return f.disposeClient(pooled)
	}

	select {
	case f.pool <- pooled:
		logging.Logger().Trace("Successfully returned client to LDAP pool")
		return nil
	default: // shouldn't happen, unless returning more clients than acquired
		logging.Logger().Debug("LDAP pool is full, disposing client")
		return f.disposeClient(pooled)
	}
}

func (f *PooledLDAPClientFactory) disposeClient(client *LDAPClientPooled) error {
	if client == nil {
		logging.Logger().Warn("Cannot dispose nil LDAP client")
		return nil
	}

	remaining := f.activeCount.Add(-1)
	logging.Logger().Debugf("Pooled LDAP client disposed; remaining: %d", remaining)
	return client.Close()
}

func (f *PooledLDAPClientFactory) acquire() (client *LDAPClientPooled, err error) {
	logging.Logger().Trace("Acquiring client from LDAP pool")

	if f.pool == nil || f.isClosing() {
		logging.Logger().Error("Cannot acquire client: LDAP pool is not initialized or closing")
		return nil, fmt.Errorf("error acquiring client: LDAP pool is not initialized or closing")
	}

	now := time.Now()
	deadline := now.Add(f.config.Pooling.Timeout)
	interval := time.Millisecond * 10

	for time.Now().Before(deadline) {
		f.wakeupPoolManager() // request more clients if needed
		select {
		case <-time.After(interval): // this avoids the blocking time.Sleep()
			continue // do nothing, except to loop again to signal the pool manager
		case client = <-f.pool:
			if !client.IsHealthy() { // NOTE: we dispose expired clients only on release
				logging.Logger().Info("Received invalid client from LDAP pool, disposing")
				_ = f.disposeClient(client)
				continue
			}
			logging.Logger().Tracef("Successfully acquired valid client from LDAP pool after %s", time.Since(now))
			return client, nil
		}
	}
	return nil, fmt.Errorf("timeout waiting for client from LDAP pool")
}

func (f *PooledLDAPClientFactory) tryAddPooledClient() (bool) {
	if f.isClosing() { return false }
	maxRetries := f.config.Pooling.Retries
	var client *LDAPClientPooled
	var err error

	sleep := f.config.Timeout / 8 // initial sleep duration, trade-off between responsiveness and overload
	for attempts := 0; attempts <= maxRetries; attempts++ {
		start := time.Now()
		client, err = f.new() // NOTE: new() adheres to config.Timeout, not config.Pooling.Timeout!
		elapsed := time.Since(start)

		if err == nil {
			// Shorten client lifetime by up to 5% to avoid mass exodus events.
			// Actual client lifetime will be slightly longer anyway as expiry is
			// checked on client release to improve availability and reduce latency.
			fuzz := time.Duration(float64(f.clientLifetime) * (0.95 + 0.05*rand.Float64()))
			client.expiresAt = time.Now().Add(fuzz)
			logging.Logger().Tracef("Created new client for LDAP pool in %s", elapsed)
			break // Successfully created client
		}

		logging.Logger().Debugf("Failed to create new client for LDAP pool (attempt %d/%d, %s): %v",
			1 + attempts, 1 + maxRetries, elapsed, err)

		if attempts < maxRetries {
			time.Sleep(sleep)
			sleep *= 2 // try again, backing off
		} else {
			logging.Logger().Warningf("Exceeded maximum retries (%d) for creating pooled LDAP client; last error: %v",
				1 + maxRetries, err)
			return false // give up
		}
	}

	select {
	case f.pool <- client:
		active := f.activeCount.Add(1)
		logging.Logger().Debugf("Added new client to LDAP pool. Active: %d, Available: %d", active, len(f.pool))
		return true
	default:
		// shouldn't happen, unless clients are returned multiple times or returned without being acquired
		logging.Logger().Warning("LDAP pool full, closing newly created client")
		_ = client.Close()
		return false
	}
}

func (f *PooledLDAPClientFactory) wakeupPoolManager() {
	select {
	case f.wakeup <- struct{}{}:
	default: // Already signaled, do nothing
	}
}

// goroutine that maintains the pool of LDAP clients
func (f *PooledLDAPClientFactory) poolManager() {
	logging.Logger().Infof("LDAP Pool Manager started: minPoolSize=%d, maxPoolSize=%d, retries=%d, timeout=%v",
		f.minPoolSize, f.maxPoolSize, f.config.Pooling.Retries, f.config.Pooling.Timeout)

	f.wakeupPoolManager() // initial wakeup to prime the pool

	for !f.isClosing() {
		select {
		case <-f.ctx.Done():
			logging.Logger().Warn("LDAP pool context cancelled, exiting...")
			return // This exits the goroutine 'ungracefully', without further cleanup

		// Signalled by one OR MORE threads that are about to OR ALREADY HAVE acquired a client
		// Also used by Close() after setting f.closing to initiate a pool cleanup and graceful exit
		case <-f.wakeup:
			if f.isClosing() { goto poolCleanup }

			// Goals:
			//  1. Ensure minimum pool size is maintained if possible - this is a soft target
			//  2. Prefer one surplus available client in addition to the request being serviced
			//  3. Avoid exceeding maximum pool size - this is a hard limit
			//  4. Reduce bursts of new connections and excessive resource usage
			//  5. Handle transient failures in creating new connections (incl. pool shutdown)

			active := f.activeCount.Load()
			available := int32(len(f.pool))
			demand := utils.Clamp( f.minPoolSize - active,	// Goal 1, maintain minimum pool size
				2 - available, f.maxPoolSize - active) 				// Goal 2 and 3
			if( demand <= 0 ) { continue }

			demand = max( demand, (f.minPoolSize + 1) / 2) 	// Goal 4, limit bursts to half minPoolSize (rounded up)
			logging.Logger().Debugf("LDAP Pool clients available: (%d/%d), Requesting: %d", available, active, demand)

			for range demand {
				if !f.tryAddPooledClient() { continue } 			// Goal 5: stop if we can't add more clients
			}
		}
	}

poolCleanup:

	logging.Logger().Debugf("LDAP pool is closing, cleaning up %d pooled clients", f.activeCount.Load())

	// no timeout here, Close() handles cancelling f.ctx if necessary
	for f.activeCount.Load() > 0 {
		select {
		case <-f.ctx.Done():
			logging.Logger().Warn("LDAP pool context cancelled during cleanup")
			f.closed <- fmt.Errorf("LDAP pool context cancelled during cleanup")
			return
		case client, ok := <-f.pool:
			if !ok {
				logging.Logger().Warn("LDAP pool channel unexpectedly closed during cleanup")
				f.closed <- fmt.Errorf("LDAP pool channel unexpectedly closed during cleanup")
				return
			}
			_ = f.disposeClient(client)
			logging.Logger().Tracef("Closed pooled LDAP client, remaining clients: %d", f.activeCount.Load())
		}
	}
	close(f.pool)
	f.pool = nil
	close(f.closed) // don't send an error value, just close to signal completion
}

func (f *PooledLDAPClientFactory) Close() (err error) {
	logging.Logger().Trace("Closing LDAP connection pool")

	if f.pool == nil {
		logging.Logger().Warn("LDAP pool is not initialized, nothing to close")
		return nil
	}

	if !f.closing.CompareAndSwap(false, true) {
		logging.Logger().Warn("LDAP pool already closing, ignoring request")
		return nil
	}

	f.wakeupPoolManager() // signal the pool boy that it's cleaning up time

	cleanupCtx, cancelCleanup := context.WithTimeout(context.Background(), f.config.Pooling.Timeout)
	defer cancelCleanup()

	select {
	case cleanupErr, hasError := <-f.closed:
		err = cleanupErr // maybe nil, if there was no error
		if !hasError {
			logging.Logger().Debug("LDAP pool cleanup complete, all clients closed")
		}
		f.closed = nil
	case <-cleanupCtx.Done(): // timeout
		f.cancel() // shoot the pool manager, if it hasn't exited gracefully
		logging.Logger().Warn("LDAP pool cleanup exceeded timeout")
		err = fmt.Errorf("LDAP pool cleanup exceeded timeout")
	}

	cleanupCtx = nil
	f.cancel = nil
	f.ctx = nil
	f.wakeup = nil

	logging.Logger().Trace("LDAP pool closed")
	return err
}

func (f *PooledLDAPClientFactory) ReadinessCheck() (err error) {
	logging.Logger().Trace("Checking if LDAP pool is ready")

	client, err := f.acquire()
	if err != nil {
		logging.Logger().Debugf("Failed to acquire client for LDAP pool readiness check: %v", err)
		return err
	}

	//
	// TBD: add a more meaningful check here?
	//

	_ = f.ReleaseClient(client)

	logging.Logger().Trace("LDAP pool is ready")
	return nil
}

// LDAPClientPooled is a decorator for the ldap.Client which handles the pooling functionality. i.e. prevents the client
// from being closed and instead relinquishes the connection back to the pool.
type LDAPClientPooled struct {
	ldap.Client
	expiresAt time.Time
	// leased bool // prevent double leasing / double returns
}

func (c *LDAPClientPooled) IsExpired() bool {
	return time.Now().After(c.expiresAt)
}

func (c *LDAPClientPooled) IsHealthy() bool {
	return c.Client != nil && !c.Client.IsClosing()
}

func (c *LDAPClientPooled) Close() error {
	if c.Client != nil {
		logging.Logger().Trace("Closing LDAP client")
		err := c.Client.Close()
		if err != nil {
			logging.Logger().Warnf("Error closing LDAP client: %v", err)
		}
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
