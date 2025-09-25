package authentication

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
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

	sleep := config.Pooling.Timeout / time.Duration( 1 + config.Pooling.Retries) // slightly smaller than the client timeout

	factory = &PooledLDAPClientFactory{
		config:        	config,
		tls:           	tlsc,
		opts:          	opts,
		dialer:        	dialer,
		sleep:         	sleep,
		// these could possibly be configured
		minPoolSize:   	int32(max(1, config.Pooling.Count / 2)),
		maxPoolSize:   	int32(config.Pooling.Count),
		clientLifetime: time.Hour,
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
	pool           chan *LDAPClientPooled  // Channel for available clients
	activeCount    int32                   // Atomic counter for active connections
	minPoolSize    int32                   // Minimum number of pool connections to maintain
	maxPoolSize    int32                   // Maximum number of pool connections
	sleep          time.Duration           // Sleep duration between retries
	clientLifetime time.Duration           // Maximum lifetime for a pooled client

	// Synchronization
	requests       chan struct{}           // Channel for client requests
	closing        int32                   // Atomic flag indicating if the pool is closing
	closed         chan struct{}           // Signals when poolManager has completed cleanup
	ctx            context.Context         // Context for cancellation
	cancel         context.CancelFunc      // Function to cancel the context
}

func (f *PooledLDAPClientFactory) isClosing() bool {
	return atomic.LoadInt32(&f.closing) > 0
}

func (f *PooledLDAPClientFactory) setClosing() bool {
	return atomic.CompareAndSwapInt32(&f.closing, 0, 1)
}

func (f *PooledLDAPClientFactory) Initialize() (err error) {

	if f.pool != nil {
		return nil
	}

	f.pool = make(chan *LDAPClientPooled, f.maxPoolSize)
	f.requests = make(chan struct{}, f.maxPoolSize) // TODO: make this 1?
	f.closed = make(chan struct{})

	f.ctx, f.cancel = context.WithCancel(context.Background())

	go f.poolManager()

	return nil
}

// GetClient opens new client using the pool.
func (f *PooledLDAPClientFactory) GetClient(opts ...LDAPClientFactoryOption) (conn ldap.Client, err error) {
	if len(opts) != 0 {
		return getLDAPClient(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts, opts...)
	}

	return f.acquire(context.Background())
}

// The new function creates a pool based client. This function is not thread safe.
func (f *PooledLDAPClientFactory) new() (pooled *LDAPClientPooled, err error) {
	var client ldap.Client

	if client, err = getLDAPClient(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts); err != nil {
		return nil, fmt.Errorf("error occurred establishing new client for the pool: %w", err)
	}

	return &LDAPClientPooled{
		Client:    client,
		expiresAt: time.Now().Add(f.clientLifetime),
	}, nil
}

// ReleaseClient returns a client using the pool or closes it.
func (f *PooledLDAPClientFactory) ReleaseClient(client ldap.Client) (err error) {
	logging.Logger().Trace("[LDAP Pool] Releasing LDAP client")

	pooled, ok := client.(*LDAPClientPooled)
	if !ok {
		logging.Logger().Debug("[LDAP Pool] Client is not a pooled client, closing directly")
		return client.Close()
	}

	if !pooled.IsHealthy() || pooled.IsExpired() {
		logging.Logger().Info("[LDAP Pool] Pooled client is not healthy or has expired, disposing")
		return f.disposeClient(pooled)
	}

	//logging.Logger().Trace("[LDAP Pool] Returning valid client to pool")
	select {
	case f.pool <- pooled:
		logging.Logger().Trace("[LDAP Pool] Successfully returned client to pool")
		return nil
	default: // shouldn't happen, unless returning more clients than acquired
		logging.Logger().Debug("[LDAP Pool] Pool is full, disposing client")
		return f.disposeClient(pooled)
	}
}

func (f *PooledLDAPClientFactory) disposeClient(client *LDAPClientPooled) error {
	if client == nil {
		logging.Logger().Warn("[LDAP Pool] Cannot dispose nil client")
		return nil
	}

	oldCount := atomic.AddInt32(&f.activeCount, -1)
	logging.Logger().Debugf("[LDAP Pool] Client disposed; decremented active count from %d to %d", oldCount+1, oldCount)

	return client.Close()
}

func (f *PooledLDAPClientFactory) acquire(ctx context.Context) (client *LDAPClientPooled, err error) {
	logging.Logger().Trace("[LDAP Pool] Acquiring LDAP client from pool")

	if f.pool == nil || f.isClosing() {
		logging.Logger().Error("[LDAP Pool] Cannot acquire client: pool is not initialized or closing")
		return nil, fmt.Errorf("error acquiring client: the pool is not initialized or closing")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, f.config.Pooling.Timeout)
	defer cancel()

	for {
		//logging.Logger().Trace("[LDAP Pool] Waiting for available client")

		select {

		case <-timeoutCtx.Done():
			logging.Logger().Errorf("[LDAP Pool] Timeout waiting for LDAP client: %v", timeoutCtx.Err())
			return nil, fmt.Errorf("timeout waiting for LDAP client: %w", timeoutCtx.Err())

		case client = <-f.pool:
			//logging.Logger().Trace("[LDAP Pool] Received client from pool")

			if !client.IsHealthy()  {
				logging.Logger().Info("[LDAP Pool] Received invalid client, disposing")
				_ = f.disposeClient(client)
				continue
			}

			logging.Logger().Trace("[LDAP Pool] Successfully acquired valid client")
			return client, nil

		default: // with the right poolsize, this shouldn't happen often
			f.wakeupPoolManager()
      time.Sleep(10 * time.Millisecond)
		}
	}
}

func (f *PooledLDAPClientFactory) tryAddPooledClient() bool {

	if f.isClosing() {
		return false
	}

	currentCount := atomic.LoadInt32(&f.activeCount)
	poolSize := len(f.pool)

	logging.Logger().Tracef("[LDAP Pool] Processing client request. Active: %d, Available: %d, Max: %d",
		currentCount, poolSize, f.maxPoolSize)

	if currentCount >= f.maxPoolSize {
		logging.Logger().Trace("[LDAP Pool] Maximum pool size reached, not creating new client")
		return false
	}

	// Try to create a new client with multiple retries
	maxRetries := f.config.Pooling.Retries
	var client *LDAPClientPooled
	var err error

	for attempts := 1; attempts <= maxRetries; attempts++ {
		client, err = f.new()
		if err == nil {
			break // Successfully created client
		}

		logging.Logger().Warnf("[LDAP Pool] Failed to create LDAP client (attempt %d/%d): %v", attempts, maxRetries, err)

		if attempts < maxRetries {
			time.Sleep(f.sleep)
		}
	}

	if err != nil {
		logging.Logger().Errorf("[LDAP Pool] Failed to create LDAP client after %d attempts", maxRetries)
		return false
	}

	select {
	case f.pool <- client:
		atomic.AddInt32(&f.activeCount, 1)
		logging.Logger().Debug("[LDAP Pool] Added new client to pool")

		return true
	default:
		logging.Logger().Debug("[LDAP Pool] Pool full, closing newly created client")
		_ = client.Close()
		return false
	}
}

func (f *PooledLDAPClientFactory) wakeupPoolManager() {
	select {
	case f.requests <- struct{}{}:
	default: // Already signaled, do nothing
	}
}

// goroutine that manages and maintains the pool of LDAP clients
func (f *PooledLDAPClientFactory) poolManager() {
	logging.Logger().Infof("[LDAP Pool] Pool manager started: minPoolSize=%d, maxPoolSize=%d, retries=%d, timeout=%v",
		f.minPoolSize, f.maxPoolSize, f.config.Pooling.Retries, f.config.Pooling.Timeout)

	logging.Logger().Trace("[LDAP Pool] Running initial pool maintenance")
	f.poolMaintenance()

	logging.Logger().Debug("[LDAP Pool] Entering main event loop")
	for ! f.isClosing(){
		select {
		case <-f.requests:
			if len(f.pool) == 0 {
				_ = f.tryAddPooledClient()
			}
		case <-time.After(time.Second * 10):
			f.poolMaintenance();
		case <-f.ctx.Done():
			logging.Logger().Warn("[LDAP Pool] Context cancelled, exit poolManager")
			return // This exits the goroutine 'ungracefully', without further cleanup
		}
	}

	_ = f.poolCleanup()
	close(f.closed) // TODO: only close if cleanup was successful?
}

func (f *PooledLDAPClientFactory) poolMaintenance() {
	currentCount := atomic.LoadInt32(&f.activeCount)
	logging.Logger().Tracef("[LDAP Pool] Periodic maintenance check. Active: %d, Available: %d, Min: %d, Max: %d",
		currentCount, len(f.pool), f.minPoolSize, f.maxPoolSize)

	// Try to directly add clients to the pool to reach minimum size
	clientsNeeded := max(0, f.minPoolSize - currentCount)
	for i := int32(0); i < clientsNeeded; i++ {
		if !f.tryAddPooledClient() {
			break // Stop if we can't add more clients
		}
	}
}

func (f *PooledLDAPClientFactory) poolCleanup() bool {
	logging.Logger().Debug("[LDAP Pool] Pool is closing")

	for {
		select {
		case client, ok := <-f.pool:
			if !ok {
				logging.Logger().Warn("[LDAP Pool] Pool channel closed during cleanup")
				return false
			}

			_ = f.disposeClient(client)
			activeCount := atomic.LoadInt32(&f.activeCount)
			if( activeCount > 0) {
				logging.Logger().Tracef("[LDAP Pool] Closed client from pool, remaining: %d", activeCount)
			} else {
				logging.Logger().Debug("[LDAP Pool] Cleanup complete, all connections closed")
				return true
			}

		case <-f.ctx.Done(): // Context was cancelled by Close(), probably after timeout
			logging.Logger().Warn("[LDAP Pool] Context cancelled during cleanup, exiting")
			return false
		}
	}
}

func (f *PooledLDAPClientFactory) Close() (err error) {
	logging.Logger().Trace("[LDAP Pool] Closing LDAP connection pool")

	if f.pool == nil {
		logging.Logger().Warn("[LDAP Pool] Pool was never initialized, nothing to close")
		return nil
	}

	if !f.setClosing() {
		logging.Logger().Warn("[LDAP Pool] Pool already closing, ignoring request")
		return nil;
	}

	go func() {
		//logging.Logger().Debug("[LDAP Pool] Signaling pool manager to start cleanup")
		f.wakeupPoolManager()

		logging.Logger().Trace("[LDAP Pool] Waiting for pool manager to complete cleanup")
		timeoutDuration := f.config.Pooling.Timeout

		select {
		case <-f.closed:
			logging.Logger().Trace("[LDAP Pool] Pool manager coroutine completed")
		case <-time.After(timeoutDuration):
			logging.Logger().Warnf("[LDAP Pool] Timed out waiting for pool manager cleanup after %v, forcing exit", timeoutDuration)
			if f.cancel != nil {
				f.cancel() // Force cancellation of the pool manager
				f.cancel = nil
			}
		}

		// Setting channels and contexts to nil helps with garbage collection
		f.pool = nil
		f.requests = nil
		f.closed = nil
		f.ctx = nil

		logging.Logger().Debug("[LDAP Pool] Pool shutdown completed")
	}()

	logging.Logger().Trace("[LDAP Pool] Pool close initiated, returning non-blocking")
	return nil
}

func (f *PooledLDAPClientFactory) IsReady() bool {
	logging.Logger().Trace("[LDAP Pool] Checking if pool is ready")

	client, err := f.acquire(context.Background())
	if err != nil {
		logging.Logger().Warnf("[LDAP Pool] Failed to acquire client for readiness check: %v", err)
		return false
	}

	// TBD: add a more meaningful check here?

	//logging.Logger().Trace("[LDAP Pool] Successfully acquired client, releasing back to pool")
	_ = f.ReleaseClient(client)

	logging.Logger().Trace("[LDAP Pool] Pool is ready")
	return true
}


// LDAPClientPooled is a decorator for the ldap.Client which handles the pooling functionality. i.e. prevents the client
// from being closed and instead relinquishes the connection back to the pool.
type LDAPClientPooled struct {
	ldap.Client
	expiresAt time.Time
	// TODO: leased bool // prevent double leasing / double returns
}

func (c *LDAPClientPooled) IsExpired() bool {
	return time.Now().After(c.expiresAt)
}

func (c *LDAPClientPooled) IsHealthy() bool {
	return c.Client != nil && !c.Client.IsClosing()
}

func (c *LDAPClientPooled) Close() error {
	if c.Client != nil {
		logging.Logger().Trace("[LDAP Pool] Closing LDAP client")
		err := c.Client.Close()
		if err != nil {
				logging.Logger().Warnf("[LDAP Pool] Error closing client: %v", err)
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
