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

	factory = &PooledLDAPClientFactory{
		log: 						logging.Logger().WithFields(map[string]any{"provider": "pooled ldap factory"}),
		config:        	config,
		tls:           	tlsc,
		opts:          	opts,
		dialer:        	dialer,
		// these could possibly be configured, or initialized differently
		minPoolSize:    int32(max(2, config.Pooling.Count/2)),
		maxPoolSize:    int32(max(2, config.Pooling.Count)),
		clientLifetime: time.Minute * 60, // this is a soft target, actual lifetime is fuzzed
	}

	return factory
}

// PooledLDAPClientFactory is a LDAPClientFactory that takes another LDAPClientFactory and pools the
// factory generated connections using a channel for thread safety.
type PooledLDAPClientFactory struct {
	log 												 *logrus.Entry
	config                       *schema.AuthenticationBackendLDAP
	tls                          *tls.Config
	opts                         []ldap.DialOpt
	dialer                       LDAPClientDialer

	// pool configuration - immutable after initialization
	minPoolSize                  int32                  // Minimum number of pool connections to maintain (soft target)
	maxPoolSize                  int32                  // Maximum number of pool connections (hard target)
	clientLifetime               time.Duration          // Maximum lifetime for a pooled client (soft target)

	// Pool management
	pool                         chan *LDAPClientPooled // Channel for available clients
	activeCount                  atomic.Int32           // Atomic counter for active connections
	clientSequence               atomic.Int32           // Atomic counter for client IDs

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
	metricsClientsUnhealthy      atomic.Int64           // Count of unhealthy client disposals
	metricsClientsCreated        atomic.Int64           // Total number of clients created
	metricsCreateTimeSum         atomic.Int64           // Sum of all client creation times (nanoseconds)
	metricsCreateFailedAttempts  atomic.Int64           // Failed attempts to create clients
	metricsCreateRetriesExceeded atomic.Int64           // Total retry failures after exhausting attempts
	metricsManagerWakeupEvents   atomic.Int64           // Count of pool manager wakeups
	metricsClientsMaxActive      atomic.Int32           // Maximum number of active clients at any time
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
		f.log.Warnf("LDAP pooling timeout (%v) may be insufficient for retry scenarios (recommended: >%v)",
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

	f.log.WithError(err).Debug("LDAP pool initialization failed")
	err = fmt.Errorf("LDAP pool initialization failed: %w", err)
	f.closed <- err // allow the Close() method to proceed immediately
	_ = f.Close()   // de-initialize
	return err
}

// thread-safe, returns nil if the pool is ready, or an error otherwise
func (f *PooledLDAPClientFactory) ReadinessCheck() (err error) {
	f.log.Trace("Checking if LDAP pool is ready")

	client, err := f.acquire()
	if err != nil {
		f.log.WithError(err).Debug("Failed to acquire client for LDAP pool readiness check")
		return err
	}

	//
	// TBD: add a more meaningful check here?
	//

	_ = f.ReleaseClient(client)

	f.log.Trace("LDAP pool is ready")
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

	id := f.clientSequence.Add(1)
	return &LDAPClientPooled{
		Client: client,
		log:    f.log.WithField("client", id),
	}, nil
}

// thread-safe, returns a client using the pool or closes it.
func (f *PooledLDAPClientFactory) ReleaseClient(client ldap.Client) (err error) {
	f.log.Trace("Releasing LDAP client")

	pooled, ok := client.(*LDAPClientPooled)
	if !ok {
		f.log.Trace("LDAP client is not pooled, closing directly")
		return client.Close()
	}

	// NOTE: clients can still be returned while the pool is closing.
	// Rationale: cleanup in poolManager() expects this and handles it gracefully.

	f_pool := f.pool // local copy for thread safety, atomic on all supported platforms
	if f_pool == nil {
		f.log.Warn("LDAP pool is not initialized or was closed, disposing client")
		return f.disposeClient(pooled)
	}

	// assert correctness
	if !pooled.leased.CompareAndSwap(true, false) {
		f.log.Warn("Pooled LDAP client was not leased, disposing to prevent double return")
		return f.disposeClient(pooled)
	}

	if pooled.IsExpired() {
		f.log.Debug("Pooled LDAP client has expired, disposing")
		return f.disposeClient(pooled)
	}

	if !pooled.IsHealthy() {
		f.log.Debug("Pooled LDAP client is not healthy, disposing")
		f.metricsClientsUnhealthy.Add(1)
		return f.disposeClient(pooled)
	}

	select {
	case f_pool <- pooled:
		f.log.Trace("Successfully returned client to LDAP pool")
		return nil
	default:
		// This shouldn't happen, but we handle it gracefully
		f.log.Warn("LDAP pool is full, disposing client")
		return f.disposeClient(pooled)
	}
}

// thread-safe, closes the client and decrements active count
func (f *PooledLDAPClientFactory) disposeClient(client *LDAPClientPooled) error {
	if client == nil {
		f.log.Warn("Cannot dispose nil LDAP client")
		return nil
	}

	remaining := f.activeCount.Add(-1)
	f.log.WithField("remainingClients", remaining).Trace("Pooled LDAP client disposed")
	f.wakeupPoolManager() // might need replacing
	return client.Close()
}

// thread-safe, returns a client using the pool or an error
func (f *PooledLDAPClientFactory) acquire() (client *LDAPClientPooled, err error) {
	f.log.Trace("Acquiring client from LDAP pool")

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
			f.log.Warn("Received already leased client from LDAP pool, disposing")
			_ = f.disposeClient(client)
			continue // for !f.isClosing()
		}

		if !client.IsHealthy() { // NOTE: we dispose expired clients only on release
			f.log.Debug("Received invalid client from LDAP pool, disposing")
			_ = f.disposeClient(client)
			f.metricsClientsUnhealthy.Add(1)
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

		f.log.WithField("elapsed", elapsed).Trace("Successfully acquired valid client from LDAP pool")
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
			f.metricsClientsCreated.Add(1)
			f.metricsCreateTimeSum.Add(time.Since(start).Nanoseconds())
			// Fuzz client lifetime by +/- 10% to avoid mass exodus events
			fuzzFactor := 0.90 + 0.2*rand.Float64() // [0.90..1.10)
			client.expiresAt = time.Now().Add(time.Duration(float64(f.clientLifetime) * fuzzFactor))

			select {
			case f_pool <- client:
				active := f.activeCount.Add(1)
				// Safely update max active clients metric
				for {
					current := f.metricsClientsMaxActive.Load()
					if active <= current || f.metricsClientsMaxActive.CompareAndSwap(current, active) {
						break
					}
				}
				f.log.WithFields(logrus.Fields{
					"active":    active,
					"available": len(f_pool),
				}).Debug("Added new client to LDAP pool")
				return true
			default: // shouldn't happen
				f.log.Warn("LDAP pool closed or full, closing newly created client")
				_ = client.Close()
				return false
			}
		}

		f.metricsCreateFailedAttempts.Add(1)
		f.log.WithError(err).WithFields(logrus.Fields{
			"attempt":     f.config.Pooling.Retries - attempts + 1,
			"maxAttempts": f.config.Pooling.Retries + 1,
			"elapsed":     time.Since(start),
		}).Debug("Failed to create new client for LDAP pool")

		if attempts == 0 {
			f.metricsCreateRetriesExceeded.Add(1)
			f.log.WithError(err).WithField("retries", f.config.Pooling.Retries).Warn("Exceeded maximum retries for creating pooled LDAP client")
			return false
		}

		f.log.WithField("sleepDuration", sleep).Trace("Sleeping before next attempt to create pooled LDAP client")
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
		err := fmt.Errorf("some components are not initialized")
		f.log.WithError(err).Error("LDAP pool manager cannot start")
		if f_closed != nil {
			f_closed <- fmt.Errorf("LDAP pool manager cannot start: %w", err)
		}
		return
	}

	f.log.WithFields(logrus.Fields{
		"minPoolSize": f.minPoolSize,
		"maxPoolSize": f.maxPoolSize,
		"retries":     f.config.Pooling.Retries,
		"timeout":     f.config.Pooling.Timeout,
	}).Info("LDAP Pool Manager started")

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
			err := fmt.Errorf("pool context cancelled")
			f.log.WithError(err).Warn("LDAP pool manager interrupted")
			f_closed <- fmt.Errorf("LDAP pool context cancelled: %w", err)
			return // This exits the goroutine 'ungracefully', without further cleanup

		case <-ticker.C:
			f.metricsReport()

		// Signalled by one OR MORE callers that initially failed to acquire a client, i.e. by the
		// time this runs, the situation might have changed already. What we *do know* is that
		// the pool is or was under pressure and we should pro-actively try to add more clients.
		// Also used by disposeClinent() after decrementing f.activeCount to allow for a pool top-up.
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

			f.log.WithFields(logrus.Fields{
				"available": available,
				"active":    active,
				"request":   request,
			}).Debug("LDAP Pool clients available, requesting increase")

			for range request {
				if !f.tryAddPooledClient(f_pool) {				// Goal 3: handle transient failures
					break
				}
			}
		}
	}

poolCleanup:

	f.log.WithField("activeClients", f.activeCount.Load()).Debug("LDAP pool is closing, cleaning up pooled clients")

	// no timeout here, Close() already handles cancelling f.ctx if necessary
	for f.activeCount.Load() > 0 {
		select {
		case <-f_ctx.Done():
			err := fmt.Errorf("pool context cancelled during cleanup")
			f.log.WithError(err).Warn("LDAP pool cleanup interrupted")
			f_closed <- fmt.Errorf("LDAP pool context cancelled during cleanup: %w", err)
			return
		case client, ok := <-f_pool:
			if !ok {
				err := fmt.Errorf("pool channel unexpectedly closed")
				f.log.WithError(err).Warn("LDAP pool cleanup interrupted")
				f_closed <- fmt.Errorf("LDAP pool channel unexpectedly closed during cleanup: %w", err)
				return
			}
			_ = f.disposeClient(client) // decrements f.activeCount
			f.log.WithField("remainingClients", f.activeCount.Load()).Trace("Closed pooled LDAP client")
		}
	}
	close(f_closed) // don't send an error value, just close to signal completion
}

// Cooperates with the pool manager goroutine to cleanup and exit gracefully
// must be called only once - subsequent calls are ignored
func (f *PooledLDAPClientFactory) Close() (err error) {
	f.log.Trace("Closing LDAP connection pool")

	if f.pool == nil {
		f.log.Warn("LDAP pool is not initialized, nothing to close")
		return nil
	}

	// signal the pool manager we're closing and prvent subsequent calls
	if !f.closing.CompareAndSwap(false, true) {
		f.log.Warn("LDAP pool already closing, ignoring request")
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
			f.log.Debug("LDAP pool cleanup complete, all clients closed")
		} else if cleanupErr != nil {
			f.log.WithError(cleanupErr).Debug("LDAP pool cleanup completed with error")
		}
	case <-cleanupCtx.Done(): // timeout
		f.cancel() // shoot the pool manager, if it hasn't exited gracefully yet
		err = fmt.Errorf("pool cleanup exceeded timeout")
		f.log.WithError(err).Warn("LDAP pool cleanup timed out")
	}

	// all thread-safe methods must capture pointers before checking for nil
	f.pool = nil
	f.wakeup = nil
	f.closed = nil
	f.ctx = nil
	f.cancel = nil
	cleanupCtx = nil

	f.log.Trace("LDAP pool closed")
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
	f.metricsClientsUnhealthy.Store(0)
	f.metricsClientsCreated.Store(0)
	f.metricsCreateTimeSum.Store(0)
	f.metricsCreateFailedAttempts.Store(0)
	f.metricsCreateRetriesExceeded.Store(0)
	f.metricsManagerWakeupEvents.Store(0)
	f.metricsClientsMaxActive.Store(0)
}

// metricsSnapshot holds calculated metrics data
// NOTE: stors may overlap with updates - this is considered acceptable
type metricsSnapshot struct {
	timestamp              	string
	durationHours          	float64
	currentActive          	int
	currentAvailable      	int
	clientsAvgEstimate      float64
	clientsMaxActive       	int32

	bin_10us               	int64 // <10us
	bin_10us_avg        		float64
	bin_100us              	int64 // <100us
	bin_1ms                	int64 // <1ms
	bin_10ms               	int64 // <10ms
	bin_100ms              	int64 // [10ms..100ms)
	bin_1s                 	int64 // [100ms..1s)
	bin_longer             	int64 // ≥1s
	totalSuccessful       	int64
	timeoutFailures        	int64

	clientsCreated         	int64
	createTimeAvgMs         float64
	createRetriesExceeded 	int64
	createFailedAttempts 		int64
	clientsUnhealthy     		int64
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

	poolActive := int(f.activeCount.Load())
	poolAvailable := len(f_pool)

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

	var createTimeAvgMs float64
	if f.metricsClientsCreated.Load() > 0 {
		createTimeAvgMs = float64(f.metricsCreateTimeSum.Load()) / float64(f.metricsClientsCreated.Load()) / 1e6
	}

	// Calculate estimated number of average active clients based on total lifetime
	var clientsAvgEstimate float64
	if duration.Hours() > 0 && f.clientLifetime > 0 {
		totalClientHours := float64(f.metricsClientsCreated.Load()) * min(f.clientLifetime.Hours(), duration.Hours())
		clientsAvgEstimate = totalClientHours / duration.Hours()
	}

	return &metricsSnapshot{
		timestamp:              now.Format("2006-01-02T15:04:05Z07:00"),
		durationHours:          duration.Hours(),
		currentActive:          poolActive,
		currentAvailable:       poolAvailable,

		bin_10us:               bin_10us,
		bin_10us_avg:        		bin_10us_avg,
		bin_100us:              bin_100us,
		bin_1ms:                bin_1ms,
		bin_10ms:               bin_10ms,
		bin_100ms:              bin_100ms,
		bin_1s:                 bin_1s,
		bin_longer:             bin_remain,
		totalSuccessful:        totalSuccessful,
		timeoutFailures:        f.metricsTimeoutFailures.Load(),

		clientsAvgEstimate:     clientsAvgEstimate,
		clientsMaxActive:       f.metricsClientsMaxActive.Load(),
		clientsCreated:         f.metricsClientsCreated.Load(),
		createTimeAvgMs:        createTimeAvgMs,
		createFailedAttempts: 	f.metricsCreateFailedAttempts.Load(),
		createRetriesExceeded:  f.metricsCreateRetriesExceeded.Load(),
		clientsUnhealthy:     	f.metricsClientsUnhealthy.Load(),
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
		"timestamp":                snapshot.timestamp,
		"metrics_duration_h":       snapshot.durationHours,
		"current_active":           snapshot.currentActive,
		"current_available":        snapshot.currentAvailable,
		// Pooled client acquisition statistics
		"bin_10us":                 snapshot.bin_10us,
		"bin_10us_avg":             snapshot.bin_10us_avg,
		"bin_100us":                snapshot.bin_100us,
		"bin_1ms":                  snapshot.bin_1ms,
		"bin_10ms":                 snapshot.bin_10ms,
		"bin_100ms":                snapshot.bin_100ms,
		"bin_1s":                   snapshot.bin_1s,
		"bin_remain":               snapshot.bin_longer,
		"bin_total":                snapshot.totalSuccessful,
		"failures_timeout":         snapshot.timeoutFailures,
		// Pool management statistics
		"clients_avg":              snapshot.clientsAvgEstimate,
		"clients_max":              snapshot.clientsMaxActive,
		"clients_created":          snapshot.clientsCreated,
		"clients_unhealthy":        snapshot.clientsUnhealthy,
		"creation_avg_ms":          snapshot.createTimeAvgMs,
		"retries_exceeded":       	snapshot.createRetriesExceeded,
		"creation_failed_attempts": snapshot.createFailedAttempts,
		"pool_drained_events":      snapshot.poolDrainedEvents,
		"manager_wakeup_events":    snapshot.managerWakeupEvents,
	}
}

// thread-safe, returns the CSV header for metrics data
func (f *PooledLDAPClientFactory) metricsCsvHeader() string {
	return "timestamp,metrics_duration_h,current_active,current_available," +
		"bin_10us,bin_100us,bin_1ms,bin_10ms,bin_100ms,bin_1s,bin_remain," +
		"bin_total,bin_10us_avg,failures_timeout," +
		"clients_max,clients_avg,clients_created,clients_unhealthy," +
		"creation_avg_ms,retries_exceeded,creation_failed_attempts," +
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
		snapshot.bin_longer,
		snapshot.totalSuccessful,
		snapshot.bin_10us_avg,
		snapshot.timeoutFailures,
		snapshot.clientsMaxActive,
		snapshot.clientsAvgEstimate,
		snapshot.clientsCreated,
		snapshot.clientsUnhealthy,
		snapshot.createTimeAvgMs,
		snapshot.createRetriesExceeded,
		snapshot.createFailedAttempts,
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
	fmt.Printf(" Max Active Clients: %d\n", snapshot.clientsMaxActive)
	fmt.Printf(" Avg Active Clients: %.2f\n", snapshot.clientsAvgEstimate)
	fmt.Printf("\n Successful Acquisitions by Time Range:\n")
	fmt.Printf("    <10us: %d   avg: %.2fμs\n", snapshot.bin_10us, snapshot.bin_10us_avg)
	fmt.Printf("   <100us: %d\n", snapshot.bin_100us)
	fmt.Printf("     <1ms: %d\n", snapshot.bin_1ms)
	fmt.Printf("    <10ms: %d\n", snapshot.bin_10ms)
	fmt.Printf("   <100ms: %d\n", snapshot.bin_100ms)
	fmt.Printf("      <1s: %d\n", snapshot.bin_1s)
	fmt.Printf("      ≥1s: %d\n", snapshot.bin_longer)
	fmt.Printf("    TOTAL: %d\n", snapshot.totalSuccessful)
	fmt.Printf("\n Failures:\n")
	fmt.Printf("   Timeout failures: %d\n", snapshot.timeoutFailures)
	fmt.Printf("\n Client Management:\n")
	fmt.Printf("            Clients created: %d\n", snapshot.clientsCreated)
	fmt.Printf("        Unhealthy disposals: %d\n", snapshot.clientsUnhealthy)
	fmt.Printf("          Avg creation time: %.2fms\n", snapshot.createTimeAvgMs)
	fmt.Printf("          Creation failures: %d\n", snapshot.createRetriesExceeded)
	fmt.Printf("   Creation failed attempts: %d\n", snapshot.createFailedAttempts)
	fmt.Printf("        Pool drained events: %d\n", snapshot.poolDrainedEvents)
	fmt.Printf("       Pool manager wakeups: %d\n", snapshot.managerWakeupEvents)
	fmt.Printf("==========================================\n\n")
}

// LDAPClientPooled is a decorator for the ldap.Client which handles the pooling functionality. i.e. prevents the client
// from being closed and instead relinquishes the connection back to the pool.
type LDAPClientPooled struct {
	ldap.Client
	log       *logrus.Entry   // structured logger with client context
	expiresAt time.Time       // expiration target, with some random fuzz
	leased    atomic.Bool     // prevent double leasing / double returns
}

func (c *LDAPClientPooled) IsExpired() bool {
	return time.Now().After(c.expiresAt)
}

func (c *LDAPClientPooled) IsHealthy() bool {
	return c.Client != nil && !c.Client.IsClosing()
}

func (c *LDAPClientPooled) Close() error {
	if c.Client != nil {
		c.log.Trace("Closing LDAP client")
		err := c.Client.Close()
		if err != nil {
			c.log.WithError(err).Warn("Error closing LDAP client")
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
