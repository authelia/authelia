package middlewares

import (
	"crypto/rand"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// TimingAttackDelayFunc describes a function for preventing timing attacks via a delay.
type TimingAttackDelayFunc func(logger *logrus.Entry, requestTime time.Time, successful *bool)

// TimingAttackDelay creates a new standard timing delay func.
func TimingAttackDelay(history int, minDelayMs float64, maxRandomMs int64, initialDelay time.Duration) TimingAttackDelayFunc {
	var (
		mutex  = &sync.Mutex{}
		cursor = 0
	)

	execDurationMovingAverage := make([]time.Duration, history)

	for i := range execDurationMovingAverage {
		execDurationMovingAverage[i] = initialDelay
	}

	return func(logger *logrus.Entry, requestTime time.Time, successful *bool) {
		successfulValue := false
		if successful != nil {
			successfulValue = *successful
		}

		execDuration := time.Since(requestTime)
		execDurationAvgMs := movingAverageIteration(execDuration, history, successfulValue, &cursor, &execDurationMovingAverage, mutex)
		actualDelayMs := calculateActualDelay(logger, execDuration, execDurationAvgMs, minDelayMs, maxRandomMs, successfulValue)
		time.Sleep(time.Duration(actualDelayMs) * time.Millisecond)
	}
}

func movingAverageIteration(value time.Duration, history int, successful bool, cursor *int, movingAvg *[]time.Duration, mutex sync.Locker) float64 {
	mutex.Lock()

	var sum int64

	for _, v := range *movingAvg {
		sum += v.Milliseconds()
	}

	if successful {
		(*movingAvg)[*cursor] = value
		*cursor = (*cursor + 1) % history
	}

	mutex.Unlock()

	return float64(sum / int64(history))
}

func calculateActualDelay(logger *logrus.Entry, execDuration time.Duration, execDurationAvgMs, minDelayMs float64, maxRandomMs int64, successful bool) (actualDelayMs float64) {
	randomDelayMs, err := rand.Int(rand.Reader, big.NewInt(maxRandomMs))
	if err != nil {
		return float64(maxRandomMs)
	}

	totalDelayMs := math.Max(execDurationAvgMs, minDelayMs) + float64(randomDelayMs.Int64())
	actualDelayMs = math.Max(totalDelayMs-float64(execDuration.Milliseconds()), 1.0)
	logger.Tracef("Timing Attack Delay successful: %t, exec duration: %d, avg execution duration: %d, random delay ms: %d, total delay ms: %d, actual delay ms: %d", successful, execDuration.Milliseconds(), int64(execDurationAvgMs), randomDelayMs.Int64(), int64(totalDelayMs), int64(actualDelayMs))

	return actualDelayMs
}
