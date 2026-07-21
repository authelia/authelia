package middlewares

import (
	"context"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/random"
)

type TimingContext interface {
	context.Context

	GetRandom() random.Provider
	GetLogger() *logrus.Entry
	RecordAuthenticationDuration(success bool, elapsed time.Duration)
}

type Delayer interface {
	Delay(ctx TimingContext, requestTime time.Time, successfulPtr *bool)
	CachedDelay(ctx TimingContext, requestTime time.Time, cachedPtr, successfulPtr *bool)
}

// NewTimingAttackDelay creates a new TimingAttackDelay with successDelay and jitter enabled and record disabled by
// default. Use the Set* methods to override these defaults.
func NewTimingAttackDelay(history int, initialDelay time.Duration) *TimingAttackDelay {
	execDurationMovingAverage := make([]int64, history)

	for i := range execDurationMovingAverage {
		execDurationMovingAverage[i] = initialDelay.Milliseconds()
	}

	return &TimingAttackDelay{
		history:                   history,
		minDelayMs:                250,
		maxJitterMs:               85,
		successDelay:              true,
		jitter:                    true,
		record:                    false,
		mutex:                     &sync.Mutex{},
		execDurationMovingAverage: execDurationMovingAverage,
	}
}

// TimingAttackDelay is used to prevent timing attacks by introducing a delay relative to a moving average of past
// request durations.
type TimingAttackDelay struct {
	history     int
	minDelayMs  float64
	maxJitterMs int64

	successDelay bool
	jitter       bool
	record       bool

	mutex                     *sync.Mutex
	cursor                    int
	execDurationMovingAverage []int64
}

func (d *TimingAttackDelay) SetMinimumDelayDuration(duration time.Duration) *TimingAttackDelay {
	ms := duration.Milliseconds()

	return d.SetMinimumDelay(float64(ms))
}

func (d *TimingAttackDelay) SetMinimumDelay(minDelayMs float64) *TimingAttackDelay {
	d.minDelayMs = minDelayMs

	return d
}

// SetSuccessDelay configures whether a delay is applied. Defaults to true.
func (d *TimingAttackDelay) SetSuccessDelay(successDelay bool) *TimingAttackDelay {
	d.successDelay = successDelay

	return d
}

// SetJitter configures whether random jitter is added to the delay. Defaults to true.
func (d *TimingAttackDelay) SetJitter(jitter bool, maxJitterMs int64) *TimingAttackDelay {
	d.jitter = jitter
	d.maxJitterMs = maxJitterMs

	return d
}

// SetRecord configures whether authentication durations are recorded via the TimingContext. Defaults to false.
func (d *TimingAttackDelay) SetRecord(record bool) *TimingAttackDelay {
	d.record = record

	return d
}

// Delay implements TimingAttackDelayFunc.
func (d *TimingAttackDelay) Delay(ctx TimingContext, requestTime time.Time, successfulPtr *bool) {
	d.CachedDelay(ctx, requestTime, nil, successfulPtr)
}

// CachedDelay implements TimingAttackDelayFunc.
func (d *TimingAttackDelay) CachedDelay(ctx TimingContext, requestTime time.Time, cachedPtr, successfulPtr *bool) {
	var cached, successful bool

	if successfulPtr != nil {
		successful = *successfulPtr
	}

	if cachedPtr != nil {
		cached = *cachedPtr
	}

	execDuration, execDurationAvgMs := d.movingAverageIteration(requestTime, cached, successful)

	if d.record {
		ctx.RecordAuthenticationDuration(successful, execDuration)
	}

	if successful && !d.successDelay {
		return
	}

	actualDelayMs := calculateActualDelay(ctx, execDuration, execDurationAvgMs, d.minDelayMs, d.maxJitterMs, d.jitter, successful)

	time.Sleep(time.Duration(actualDelayMs) * time.Millisecond)
}

func (d *TimingAttackDelay) movingAverageIteration(requestTime time.Time, cached, successful bool) (execDuration time.Duration, execDurationAvgMs float64) {
	d.mutex.Lock()

	var sum int64

	for _, v := range d.execDurationMovingAverage {
		sum += v
	}

	execDuration = time.Since(requestTime)

	if successful && !cached {
		d.execDurationMovingAverage[d.cursor] = execDuration.Milliseconds()
		d.cursor = (d.cursor + 1) % d.history
	}

	d.mutex.Unlock()

	return execDuration, float64(sum) / float64(d.history)
}

func calculateActualDelay(ctx TimingContext, execDuration time.Duration, execDurationAvgMs, minDelayMs float64, maxRandomMs int64, jitter, successful bool) (actualDelayMs float64) {
	var jitterMs *big.Int

	if jitter {
		jitterMs, _ = ctx.GetRandom().IntErr(big.NewInt(maxRandomMs))
	}

	if jitterMs == nil {
		jitterMs = big.NewInt(0)
	}

	totalDelayMs := math.Max(execDurationAvgMs, minDelayMs) + float64(jitterMs.Int64())
	actualDelayMs = math.Max(totalDelayMs-float64(execDuration.Milliseconds()), 1.0)

	ctx.GetLogger().Tracef("Timing Attack Delay successful: %t, exec duration: %d, avg execution duration: %d, random delay ms: %d, total delay ms: %d, actual delay ms: %d", successful, execDuration.Milliseconds(), int64(execDurationAvgMs), jitterMs.Int64(), int64(totalDelayMs), int64(actualDelayMs))

	return actualDelayMs
}
