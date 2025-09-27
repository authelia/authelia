package authentication

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/rand"
	"net"
	"os"
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
		clientLifetime: time.Minute * 10, // this is a soft target, actual lifetime is fuzzed
	}

	return factory
}

// PooledLDAPClientFactory is a LDAPClientFactory that takes another LDAPClientFactory and pools the
// factory generated connections using a channel for thread safety.
type PooledLDAPClientFactory struct {
	config                       *schema.AuthenticationBackendLDAP
	tls                          *tls.Config
	opts                         []ldap.DialOpt
	dialer                       LDAPClientDialer

	// Pool management
	pool                         chan *LDAPClientPooled // Channel for available clients
	activeCount                  atomic.Int32           // Atomic counter for active connections
	minPoolSize                  int32                  // Minimum number of pool connections to maintain (soft target)
	maxPoolSize                  int32                  // Maximum number of pool connections (hard target)
	clientLifetime               time.Duration          // Maximum lifetime for a pooled client (soft target)

	// Synchronization
	wakeup                       chan struct{}          // Channel for client requests
	closing                      atomic.Bool            // Atomic flag indicating if the pool is closing
	closed                       chan error             // Signals when poolManager has completed cleanup
	ctx                          context.Context        // Context for cancellation
	cancel                       context.CancelFunc     // Function to cancel the context

	// Metrics fields
	metricsStartTime             time.Time              // When metrics collection started or was last reset
	metricsSuccessfulBin0        atomic.Int64           // Successful acquires [0..10us)
	metricsSuccessfulBin1        atomic.Int64           // Successful acquires [10us..100us)
	metricsSuccessfulBin2        atomic.Int64           // Successful acquires [100us..1ms]
	metricsSuccessfulBin3        atomic.Int64           // Successful acquires [1ms..10ms)
	metricsSuccessfulBin4        atomic.Int64           // Successful acquires [10ms..100ms)
	metricsSuccessfulBin5        atomic.Int64           // Successful acquires [100ms..1s)
	metricsSuccessfulBin6        atomic.Int64           // Successful acquires [1s..timeout)
	metricsBin0TimeSum           atomic.Int64           // Sum of first-try acquisition times (nanoseconds)
	metricsTimeoutFailures       atomic.Int64           // Failed acquires due to timeout
	metricsPoolDrainedEvents     atomic.Int64           // Count of times pool was found empty during acquire
	metricsUnhealthyDisposals    atomic.Int64           // Count of unhealthy client disposals
	metricsCreateCount           atomic.Int64           // Total number of clients created
	metricsCreateTimeSum         atomic.Int64           // Sum of all client creation times (nanoseconds)
	metricsCreateFailAttempts    atomic.Int64           // Failed attempts to create clients
	metricsCreateRetriesExceeded atomic.Int64           // Total retry failures after exhausting attempts
	metricsManagerWakeupEvents   atomic.Int64           // Count of pool manager wakeups
	metricsMaxActiveClients      atomic.Int32           // Maximum number of active clients at any time
}

func (f *PooledLDAPClientFactory) isClosing() bool {
	return f.closing.Load()
}

// initializes the pool and starts the pool manager goroutine
// must be called only once
func (f *PooledLDAPClientFactory) Initialize() (err error) {
	if f.pool != nil {
		return fmt.Errorf("LDAP pool is already initialized")
	}

	// Rationale: +2 to account for initial attempt and backoff intervals
	minRecommendedTimeout := f.config.Timeout * time.Duration(f.config.Pooling.Retries+2)
	if f.config.Pooling.Timeout <= minRecommendedTimeout {
		logging.Logger().Warnf("LDAP pooling timeout (%v) may be insufficient for retry scenarios (recommended: >%v)",
			f.config.Pooling.Timeout, minRecommendedTimeout)
	}

	f.pool = make(chan *LDAPClientPooled, f.maxPoolSize)
	f.wakeup = make(chan struct{}, 1)
	f.closed = make(chan error, 1)

	f.ctx, f.cancel = context.WithCancel(context.Background())

	f.metricsReset()

	// prime the pool to test connectivity and run a quick self-check
	if !f.tryAddPooledClient(f.pool) {
		err = fmt.Errorf("LDAP pool couldn't acquire initial client")
	} else if rerr := f.ReadinessCheck(); rerr != nil {
		err = rerr // failed readiness check
	} else {
		go f.poolManager() // transfer responsibility to the pool manager
		return nil
	}

	logging.Logger().Debugf("LDAP pool initialization failed: %v", err)
	err = fmt.Errorf("LDAP pool initialization failed: %w", err)
	f.closed <- err // allow the Close() method to proceed immediately
	_ = f.Close()   // de-initialize
	return err
}

// thread-safe, returns nil if the pool is ready, or an error otherwise
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

// thread-safe, opens new client using the pool.
func (f *PooledLDAPClientFactory) GetClient(opts ...LDAPClientFactoryOption) (conn ldap.Client, err error) {
	if len(opts) != 0 {
		return getLDAPClient(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts, opts...)
	}

	return f.acquire()
}

// TODO: Clarify why this method is not thread-safe - or does this only apply to the returnd client?

// The new function creates a pool based client. This function is not thread safe.
// NOTE: this function adheres to config.Timeout, not config.Pooling.Timeout!
func (f *PooledLDAPClientFactory) new() (pooled *LDAPClientPooled, err error) {
	client, err := getLDAPClient(f.config.Address.String(), f.config.User, f.config.Password, f.config.Timeout, f.dialer, f.tls, f.config.StartTLS, f.opts)
	if err != nil {
		return nil, err
	}

	return &LDAPClientPooled{Client: client}, nil
}

// thread-safe, returns a client using the pool or closes it.
func (f *PooledLDAPClientFactory) ReleaseClient(client ldap.Client) (err error) {
	logging.Logger().Trace("Releasing LDAP client")

	pooled, ok := client.(*LDAPClientPooled)
	if !ok {
		logging.Logger().Trace("LDAP client is not pooled, closing directly")
		return client.Close()
	}

	// NOTE: clients can still be returned while the pool is closing.
	// Rationale: cleanup in poolManager() expects this and handles it gracefully.

	f_pool := f.pool // local copy for thread safety, atomic on all supported platforms
	if f_pool == nil {
		logging.Logger().Warn("LDAP pool is not initialized or was closed, disposing client")
		return f.disposeClient(pooled)
	}

	// assert correctness
	if !pooled.leased.CompareAndSwap(true, false) {
		logging.Logger().Warn("Pooled LDAP client was not leased, disposing to prevent double return")
		return f.disposeClient(pooled)
	}

	if pooled.IsExpired() {
		logging.Logger().Debug("Pooled LDAP client has expired, disposing")
		return f.disposeClient(pooled)
	}

	if !pooled.IsHealthy() {
		logging.Logger().Debug("Pooled LDAP client is not healthy, disposing")
		f.metricsUnhealthyDisposals.Add(1)
		return f.disposeClient(pooled)
	}

	select {
	case f_pool <- pooled:
		logging.Logger().Trace("Successfully returned client to LDAP pool")
		return nil
	default:
		// This shouldn't happen, but we handle it gracefully
		logging.Logger().Warn("LDAP pool is full, disposing client")
		return f.disposeClient(pooled)
	}
}

// thread-safe, closes the client and decrements active count
func (f *PooledLDAPClientFactory) disposeClient(client *LDAPClientPooled) error {
	if client == nil {
		logging.Logger().Warn("Cannot dispose nil LDAP client")
		return nil
	}

	remaining := f.activeCount.Add(-1)
	logging.Logger().Tracef("Pooled LDAP client disposed; remaining: %d", remaining)
	f.wakeupPoolManager() // might need replacing
	return client.Close()
}

// thread-safe, returns a client using the pool or an error
func (f *PooledLDAPClientFactory) acquire() (client *LDAPClientPooled, err error) {
	logging.Logger().Trace("Acquiring client from LDAP pool")

	f_pool := f.pool // local copy for thread safety, atomic on all supported platforms
	if f_pool == nil {
		return nil, fmt.Errorf("error acquiring client: LDAP pool is unitialized or closed")
	}

	const interval = time.Millisecond * 10 // updat metrics calculations if changing this value
	wakeUps := int32(0)                    // metrics: count how many times we've woken up the pool manager
	poolDrained := false                   // metrics: was the pool empty when we started waiting?
	start := time.Now()
	deadline := start.Add(f.config.Pooling.Timeout) // TODO: use ( f.config.Timeout + 1 ) ???

	for !f.isClosing() {

		select {
		case client = <-f_pool: // fast path
			goto check_client // aka break

		default: // slow path - pool is empty right now
			poolDrained = true
			f.wakeupPoolManager() // request more clients if needed
			select {
			case client = <-f_pool:
				goto check_client // aka break

			case <-time.After(interval):
				if time.Now().After(deadline) {
					f.metricsTimeoutFailures.Add(1)
					return nil, fmt.Errorf("timeout waiting for client from LDAP pool")
				}
				wakeUps++
				continue // for !f.isClosing()
			}
		}

	check_client:

		// Rationale: Use the opportunity yo proactively wake up the pool manager if we're running low.
		// Avoid excessive wake-ups if the pool size is small and drained pools are common.
		if len(f_pool) <= 1 && f.maxPoolSize > 5 {
			f.wakeupPoolManager()
		} //

		// asset correctness
		if !client.leased.CompareAndSwap(false, true) {
			logging.Logger().Warn("Received already leased client from LDAP pool, disposing")
			_ = f.disposeClient(client)
			continue // for !f.isClosing()
		}

		if !client.IsHealthy() { // NOTE: we dispose expired clients only on release
			logging.Logger().Debug("Received invalid client from LDAP pool, disposing")
			_ = f.disposeClient(client)
			f.metricsUnhealthyDisposals.Add(1)
			continue // for !f.isClosing()
		}

		// NOTE: we do NOT check for expiration here - it's a randomized "soft" target.
		// Rationale: It's acceptaple to use a client beyond their expiration time.
		// Expiration is therefore checked on release only to avoid potential starvation.

		// Collect successful acquisition metrics
		elapsed := time.Since(start)

		if wakeUps == 0 {
			if elapsed < time.Microsecond*10 {
				f.metricsSuccessfulBin0.Add(1) // < 10us
				f.metricsBin0TimeSum.Add(elapsed.Nanoseconds())
			} else if elapsed < time.Microsecond*100 {
				f.metricsSuccessfulBin1.Add(1) // < 100us
			} else if elapsed < time.Millisecond {
				f.metricsSuccessfulBin2.Add(1) // < 1ms
			} else {
				f.metricsSuccessfulBin3.Add(1) // < 10ms
			}
		} else if wakeUps < 10 {
			f.metricsSuccessfulBin4.Add(1) // [10ms..100ms)
		} else if wakeUps < 100 {
			f.metricsSuccessfulBin5.Add(1) // [100ms..1000ms)
		} else {
			f.metricsSuccessfulBin6.Add(1) // > 1000ms
		}
		if poolDrained {
			f.metricsPoolDrainedEvents.Add(1)
		}

		logging.Logger().Tracef("Successfully acquired valid client from LDAP pool after %s", elapsed)
		return client, nil
	}
	return nil, fmt.Errorf("LDAP pool is closing, cannot acquire client")
}

// thread safe, but intended to be called only by Initialize() or poolManager()
func (f *PooledLDAPClientFactory) tryAddPooledClient(f_pool chan *LDAPClientPooled) bool {
	attempts := f.config.Pooling.Retries

	// Rationale: 1/8 + 1/4 + 1/2 < config.Timeout with some margin
	// trade-off between responsiveness and overload with 3 retries
	sleep := f.config.Timeout / 8

	for !f.isClosing() {
		start := time.Now()
		client, err := f.new() // NOTE: new() adheres to config.Timeout, not config.Pooling.Timeout!
		if err == nil {
			f.metricsCreateCount.Add(1)
			f.metricsCreateTimeSum.Add(time.Since(start).Nanoseconds())
			// Fuzz client lifetime by +/- 10% to avoid mass exodus events
			fuzzFactor := 0.90 + 0.2*rand.Float64() // [0.90..1.10)
			client.expiresAt = time.Now().Add(time.Duration(float64(f.clientLifetime) * fuzzFactor))

			select {
			case f_pool <- client:
				active := f.activeCount.Add(1)
				// Safely update max active clients metric
				for {
					current := f.metricsMaxActiveClients.Load()
					if active <= current || f.metricsMaxActiveClients.CompareAndSwap(current, active) {
						break
					}
				}
				logging.Logger().Debugf("Added new client to LDAP pool. Active: %d, Available: %d", active, len(f_pool))
				return true
			default: // shouldn't happen
				logging.Logger().Warn("LDAP pool closed or full, closing newly created client")
				_ = client.Close()
				return false
			}
		}

		f.metricsCreateFailAttempts.Add(1)
		logging.Logger().Debugf("Failed to create new client for LDAP pool. Attempt %d/%d, Elapsed: %s: %v",
			f.config.Pooling.Retries-attempts+1, f.config.Pooling.Retries+1, time.Since(start), err)

		if attempts == 0 {
			f.metricsCreateRetriesExceeded.Add(1)
			logging.Logger().Warnf("Exceeded maximum attempts (%d) for creating pooled LDAP client; last error: %v",
				f.config.Pooling.Retries+1, err)
			return false
		}

		logging.Logger().Tracef("Sleeping %s before next attempt to create pooled LDAP client", sleep)
		time.Sleep(sleep)
		sleep *= 2
		attempts--
	}
	return false // pool is closing
}

// thread-safe, does not block
func (f *PooledLDAPClientFactory) wakeupPoolManager() {
	f_wakeup := f.wakeup // local copy for thread safety, atomic on all supported platforms
	if f_wakeup == nil {
		return
	}
	select {
	case f_wakeup <- struct{}{}:
	default: // Already signaled, do nothing
	}
}

// goroutine that maintains the pool of LDAP clients, started by Initialize()
// TODO: consider a more generalized backoff strategy here, not based on individual connection attempts
func (f *PooledLDAPClientFactory) poolManager() {
	// local captures for thread safety, atomic on all supported platforms
	f_ctx, f_pool, f_wakeup, f_closed := f.ctx, f.pool, f.wakeup, f.closed
	if f_pool == nil || f_ctx == nil || f_wakeup == nil || f_closed == nil {
		logging.Logger().Error("LDAP pool manager cannot start: some components are not initialized")
		if f_closed != nil {
			f_closed <- fmt.Errorf("LDAP pool manager cannot start: some components are not initialized")
		}
		return
	}

	logging.Logger().Infof("LDAP Pool Manager started: minPoolSize=%d, maxPoolSize=%d, retries=%d, timeout=%v",
		f.minPoolSize, f.maxPoolSize, f.config.Pooling.Retries, f.config.Pooling.Timeout)

	f.wakeupPoolManager() // initial wakeup to prime the pool

	// Set up metrics reporting
	reportingInterval := time.Hour * 1
	if _, exists := os.LookupEnv("INFRA_HOST"); exists {
		reportingInterval = time.Second * 180
	}
	ticker := time.NewTicker(reportingInterval)
	defer ticker.Stop()

	for !f.isClosing() {
		select {
		case <-f_ctx.Done():
			logging.Logger().Warn("LDAP pool context cancelled")
			f_closed <- fmt.Errorf("LDAP pool context cancelled")
			return // This exits the goroutine 'ungracefully', without further cleanup

		case <-ticker.C:
			f.metricsReport()

		// Signalled by one OR MORE callers that initially failed to acquire a client, i.e. by the
		// time this runs, the situation might have changed already. What we *do know* is that
		// the pool is or was under pressure and we should pro-actively try to add more clients.
		// Also used by disposeClinent() after decrementing f.activeCount to initiate a pool refill if needed.
		// Also used by Close() after setting f.closing to initiate a pool cleanup and graceful exit
		case <-f_wakeup:
			f.metricsManagerWakeupEvents.Add(1)
			if f.isClosing() {
				goto poolCleanup
			} // fast path if we're closing

			// Goals to calculate request heuristics - in order of priority:
			//  1. Avoid exceeding maximum pool size - this is a hard limit
			//  2. Reduce bursts of new connections and excessive resource usage
			//  3. Handle transient failures in creating new connections (incl. pool shutdown)
			//  4. Ensure minimum pool size is maintained if possible - this is a soft target
			//  5. Prefer one surplus available client in addition to the request being serviced

			active := f.activeCount.Load()  // momentary - can only decrement outside of poolManager()
			available := int32(len(f_pool)) // momentary - no influence on maintaining minPoolSize

			request := f.minPoolSize - active            // Goal 4: ensure minimum pool size
			request = max(request, 2-available)          // Goal 5: prefer one surplus if pressure is high
			request = min(request, f.maxPoolSize-active) // Goal 1: never exceed maxPoolSize - hard limit
			if request <= 0 {
				continue
			}

			burstLimit := max(2, (f.minPoolSize+1)/2) 	// minPoolSize/2 rounded up, allow at least 2
			request = min(request, burstLimit)        	// Goal 2: limit bursts

			logging.Logger().Debugf("LDAP Pool clients available: (%d/%d), requesting +%d increase", available, active, request)

			for range request {
				if !f.tryAddPooledClient(f_pool) {				// Goal 3: handle transient failures
					break
				}
			}
		}
	}

poolCleanup:

	logging.Logger().Debugf("LDAP pool is closing, cleaning up %d pooled clients", f.activeCount.Load())

	// no timeout here, Close() already handles cancelling f.ctx if necessary
	for f.activeCount.Load() > 0 {
		select {
		case <-f_ctx.Done():
			logging.Logger().Warn("LDAP pool context cancelled during cleanup")
			f_closed <- fmt.Errorf("LDAP pool context cancelled during cleanup")
			return
		case client, ok := <-f_pool:
			if !ok {
				logging.Logger().Warn("LDAP pool channel unexpectedly closed during cleanup")
				f_closed <- fmt.Errorf("LDAP pool channel unexpectedly closed during cleanup")
				return
			}
			_ = f.disposeClient(client) // decrements f.activeCount
			logging.Logger().Tracef("Closed pooled LDAP client, remaining clients: %d", f.activeCount.Load())
		}
	}
	close(f_closed) // don't send an error value, just close to signal completion
}

// Cooperates with the pool manager goroutine to cleanup and exit gracefully
// must be called only once - subsequent calls are ignored
func (f *PooledLDAPClientFactory) Close() (err error) {
	logging.Logger().Trace("Closing LDAP connection pool")

	if f.pool == nil {
		logging.Logger().Warn("LDAP pool is not initialized, nothing to close")
		return nil
	}

	// signal the pool manager we're closing and prvent subsequent calls
	if !f.closing.CompareAndSwap(false, true) {
		logging.Logger().Warn("LDAP pool already closing, ignoring request")
		return nil
	}

	f.wakeupPoolManager() // hey, pool boy, cleaning-up time

	// Rationale: pool manager might be creating a new() client right now which could
	// block for a bit longer than f.config.Timeout under worst case conditions. We
	// therefore give it enough time to finish and then cleanup gracefully.
	// A minimum of 1s is enforced to prevent immediate context cancellations.
	timeout := max(time.Second, 2*f.config.Timeout)
	cleanupCtx, cancelCleanup := context.WithTimeout(context.Background(), timeout)
	defer cancelCleanup()

	select {
	case cleanupErr, hasError := <-f.closed:
		err = cleanupErr // maybe nil, if there was no error
		if !hasError {
			logging.Logger().Debug("LDAP pool cleanup complete, all clients closed")
		}
	case <-cleanupCtx.Done(): // timeout
		f.cancel() // shoot the pool manager, if it hasn't exited gracefully yet
		logging.Logger().Warn("LDAP pool cleanup exceeded timeout")
		err = fmt.Errorf("LDAP pool cleanup exceeded timeout")
	}

	// all thread-safe methods must capture pointers before checking for nil
	f.pool = nil
	f.wakeup = nil
	f.closed = nil
	f.ctx = nil
	f.cancel = nil
	cleanupCtx = nil

	logging.Logger().Trace("LDAP pool closed")
	return err
}

// thread-safe, resets all metrics counters and sets the start time to now
func (f *PooledLDAPClientFactory) metricsReset() {
	f.metricsStartTime = time.Now()
	f.metricsSuccessfulBin0.Store(0)
	f.metricsSuccessfulBin1.Store(0)
	f.metricsSuccessfulBin2.Store(0)
	f.metricsSuccessfulBin3.Store(0)
	f.metricsSuccessfulBin4.Store(0)
	f.metricsSuccessfulBin5.Store(0)
	f.metricsSuccessfulBin6.Store(0)
	f.metricsBin0TimeSum.Store(0)
	f.metricsTimeoutFailures.Store(0)
	f.metricsPoolDrainedEvents.Store(0)
	f.metricsBin0TimeSum.Store(0)
	f.metricsUnhealthyDisposals.Store(0)
	f.metricsCreateCount.Store(0)
	f.metricsCreateTimeSum.Store(0)
	f.metricsCreateFailAttempts.Store(0)
	f.metricsCreateRetriesExceeded.Store(0)
	f.metricsManagerWakeupEvents.Store(0)
	f.metricsMaxActiveClients.Store(0)
}

// metricsSnapshot holds calculated metrics data
// NOTE: stors may overlap with updates - this is considered acceptable
type metricsSnapshot struct {
	timestamp              	string
	durationHours          	float64
	currentActive          	int
	currentAvailable      	int
	totalSuccessful       	int64
	bin_10us_avg        		float64
	avgCreationMs          	float64
	avgActive              	float64
	bin_10us               	int64 // <10us
	bin_100us              	int64 // <100us
	bin_1ms                	int64 // <1ms
	bin_10ms               	int64 // <10ms
	bin_100ms              	int64 // [10ms..100ms)
	bin_1s                 	int64 // [100ms..1s)
	bin_remain             	int64 // ≥1s
	timeoutFailures        	int64
	maxActiveClients       	int32
	clientsCreated         	int64
	unhealthyDisposals     	int64
	creationFailures      	int64
	creationFailedAttempts 	int64
	poolDrainedEvents      	int64
	managerWakeupEvents    	int64
}

// thread-safe, creates a snapshot of current metrics with all calculations performed
// NOTE: loads may overlap with updates - this is considered acceptable
func (f *PooledLDAPClientFactory) metricsCalculateSnapshot() *metricsSnapshot {
	f_pool := f.pool // local copy for thread safety, atomic on all supported platforms
	if f_pool == nil {
		return nil
	}

	now := time.Now()
	duration := now.Sub(f.metricsStartTime)

	active := int(f.activeCount.Load())
	available := len(f_pool)

	bin_10us := f.metricsSuccessfulBin0.Load()
	bin_100us := f.metricsSuccessfulBin1.Load()
	bin_1ms := f.metricsSuccessfulBin2.Load()
	bin_10ms := f.metricsSuccessfulBin3.Load()
	bin_100ms := f.metricsSuccessfulBin4.Load()
	bin_1s := f.metricsSuccessfulBin5.Load()
	bin_remain := f.metricsSuccessfulBin6.Load()

	totalSuccessful := bin_10us + bin_100us + bin_1ms + bin_10ms + bin_100ms + bin_1s + bin_remain

	var bin_10us_avg float64
	if bin_10us > 0 {
		bin_10us_avg = float64(f.metricsBin0TimeSum.Load()) / float64(bin_10us) / 1e3
	}

	var avgCreationMs float64
	if f.metricsCreateCount.Load() > 0 {
		avgCreationMs = float64(f.metricsCreateTimeSum.Load()) / float64(f.metricsCreateCount.Load()) / 1e6
	}

	// Calculate average number of active clients based on total lifetime
	var avgActive float64
	if duration.Hours() > 0 && f.clientLifetime > 0 {
		totalClientHours := float64(f.metricsCreateCount.Load()) * min(f.clientLifetime.Hours(), duration.Hours())
		avgActive = totalClientHours / duration.Hours()
	}

	return &metricsSnapshot{
		timestamp:              now.Format("2006-01-02T15:04:05Z07:00"),
		durationHours:          duration.Hours(),
		currentActive:          active,
		currentAvailable:       available,
		totalSuccessful:        totalSuccessful,
		bin_10us_avg:        bin_10us_avg,
		avgCreationMs:          avgCreationMs,
		avgActive:              avgActive,
		bin_10us:               bin_10us,
		bin_100us:              bin_100us,
		bin_1ms:                bin_1ms,
		bin_10ms:               bin_10ms,
		bin_100ms:              bin_100ms,
		bin_1s:                 bin_1s,
		bin_remain:             bin_remain,
		timeoutFailures:        f.metricsTimeoutFailures.Load(),
		maxActiveClients:       f.metricsMaxActiveClients.Load(),
		clientsCreated:         f.metricsCreateCount.Load(),
		unhealthyDisposals:     f.metricsUnhealthyDisposals.Load(),
		creationFailures:       f.metricsCreateRetriesExceeded.Load(),
		creationFailedAttempts: f.metricsCreateFailAttempts.Load(),
		poolDrainedEvents:      f.metricsPoolDrainedEvents.Load(),
		managerWakeupEvents:    f.metricsManagerWakeupEvents.Load(),
	}
}

// thread-safe, returns detailed metrics including averages
func (f *PooledLDAPClientFactory) Metrics() map[string]interface{} {
	snapshot := f.metricsCalculateSnapshot()
	if snapshot == nil {
		return map[string]interface{}{"error": "pool not initialized"}
	}

	return map[string]interface{}{
		"timestamp":                 snapshot.timestamp,
		"metrics_duration_h":        snapshot.durationHours,
		"current_active":            snapshot.currentActive,
		"current_available":         snapshot.currentAvailable,
		// Pooled client acquisition statistics
		"bin_10us":                  snapshot.bin_10us,
		"bin_100us":                 snapshot.bin_100us,
		"bin_1ms":                   snapshot.bin_1ms,
		"bin_10ms":                  snapshot.bin_10ms,
		"bin_100ms":                 snapshot.bin_100ms,
		"bin_1s":                    snapshot.bin_1s,
		"bin_remain":                snapshot.bin_remain,
		"bin_total":                 snapshot.totalSuccessful,
		"bin_10us_avg":              snapshot.bin_10us_avg,
		"failures_timeout":          snapshot.timeoutFailures,
		// Pool management statistics
		"clients_max":               snapshot.maxActiveClients,
		"clients_avg":               snapshot.avgActive,
		"clients_created":           snapshot.clientsCreated,
		"clients_unhealthy":         snapshot.unhealthyDisposals,
		"creation_avg_ms":           snapshot.avgCreationMs,
		"creation_failures":         snapshot.creationFailures,
		"creation_failed_attempts":  snapshot.creationFailedAttempts,
		"pool_drained_events":       snapshot.poolDrainedEvents,
		"manager_wakeup_events":     snapshot.managerWakeupEvents,
	}
}

// thread-safe, returns the CSV header for metrics data
func (f *PooledLDAPClientFactory) metricsCsvHeader() string {
	return "timestamp,metrics_duration_h,current_active,current_available," +
		"bin_10us,bin_100us,bin_1ms,bin_10ms,bin_100ms,bin_1s,bin_remain," +
		"bin_total,bin_10us_avg,failures_timeout," +
		"clients_max,clients_avg,clients_created,clients_unhealthy," +
		"creation_avg_ms,creation_failures,creation_failed_attempts," +
		"pool_drained_events,manager_wakeup_events"
}

// thread-safe, returns the current metrics as a CSV data row
func (f *PooledLDAPClientFactory) metricsCsvData() string {
	snapshot := f.metricsCalculateSnapshot()
	if snapshot == nil {
		return "error, pool not initialized"
	}

	return fmt.Sprintf("%s,%.2f,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%.2f,%d,%d,%.2f,%d,%d,%.2f,%d,%d,%d,%d",
		snapshot.timestamp,
		snapshot.durationHours,
		snapshot.currentActive,
		snapshot.currentAvailable,
		snapshot.bin_10us,
		snapshot.bin_100us,
		snapshot.bin_1ms,
		snapshot.bin_10ms,
		snapshot.bin_100ms,
		snapshot.bin_1s,
		snapshot.bin_remain,
		snapshot.totalSuccessful,
		snapshot.bin_10us_avg,
		snapshot.timeoutFailures,
		snapshot.maxActiveClients,
		snapshot.avgActive,
		snapshot.clientsCreated,
		snapshot.unhealthyDisposals,
		snapshot.avgCreationMs,
		snapshot.creationFailures,
		snapshot.creationFailedAttempts,
		snapshot.poolDrainedEvents,
		snapshot.managerWakeupEvents,
	)
}

// thread-safe, metricsReport prints a formatted metrics report to stdout
func (f *PooledLDAPClientFactory) metricsReport() {
	fmt.Printf("\n======== LDAP Pool Metrics Report ========\n")

	snapshot := f.metricsCalculateSnapshot()
	if snapshot == nil {
		fmt.Printf("ERROR: Pool not initialized\n")
		fmt.Printf("==========================================\n\n")
		return
	}

	fmt.Printf(" Report Time: %s\n", time.Now().Format("2006-01-02 15:04:05 MST"))
	fmt.Printf(" Collection Duration: %.2f hours\n", snapshot.durationHours)
	fmt.Printf(" Current State: %d active, %d available\n", snapshot.currentActive, snapshot.currentAvailable)
	fmt.Printf(" Max Active Clients: %d\n", snapshot.maxActiveClients)
	fmt.Printf(" Avg Active Clients: %.2f\n", snapshot.avgActive)
	fmt.Printf("\n Successful Acquisitions by Time Range:\n")
	fmt.Printf("    <10us: %d   avg: %.2fμs\n", snapshot.bin_10us, snapshot.bin_10us_avg)
	fmt.Printf("   <100us: %d\n", snapshot.bin_100us)
	fmt.Printf("     <1ms: %d\n", snapshot.bin_1ms)
	fmt.Printf("    <10ms: %d\n", snapshot.bin_10ms)
	fmt.Printf("   <100ms: %d\n", snapshot.bin_100ms)
	fmt.Printf("      <1s: %d\n", snapshot.bin_1s)
	fmt.Printf("      ≥1s: %d\n", snapshot.bin_remain)
	fmt.Printf("    TOTAL: %d\n", snapshot.totalSuccessful)
	fmt.Printf(" \nFailures:\n")
	fmt.Printf("   Timeout failures: %d\n", snapshot.timeoutFailures)
	fmt.Printf(" \nClient Management:\n")
	fmt.Printf("            Clients created: %d\n", snapshot.clientsCreated)
	fmt.Printf("        Unhealthy disposals: %d\n", snapshot.unhealthyDisposals)
	fmt.Printf("          Avg creation time: %.2f ms\n", snapshot.avgCreationMs)
	fmt.Printf("          Creation failures: %d\n", snapshot.creationFailures)
	fmt.Printf("   Creation failed attempts: %d\n", snapshot.creationFailedAttempts)
	fmt.Printf("        Pool drained events: %d\n", snapshot.poolDrainedEvents)
	fmt.Printf("       Pool manager wakeups: %d\n", snapshot.managerWakeupEvents)
	fmt.Printf("==========================================\n\n")
}

// LDAPClientPooled is a decorator for the ldap.Client which handles the pooling functionality. i.e. prevents the client
// from being closed and instead relinquishes the connection back to the pool.
type LDAPClientPooled struct {
	ldap.Client
	expiresAt time.Time 	// expiration target, with some random fuzz
	leased 		atomic.Bool // prevent double leasing / double returns
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
