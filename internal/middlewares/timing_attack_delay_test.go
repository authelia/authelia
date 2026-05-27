package middlewares

import (
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/random"
)

func TestTimingAttackDelayAverages(t *testing.T) {
	execDuration := time.Millisecond * 500
	oneSecond := time.Millisecond * 1000
	d := &TimingAttackDelay{
		history:                   10,
		mutex:                     &sync.Mutex{},
		execDurationMovingAverage: []int64{oneSecond.Milliseconds(), oneSecond.Milliseconds(), oneSecond.Milliseconds(), oneSecond.Milliseconds(), oneSecond.Milliseconds(), oneSecond.Milliseconds(), oneSecond.Milliseconds(), oneSecond.Milliseconds(), oneSecond.Milliseconds(), oneSecond.Milliseconds()},
	}
	_, avgExecDuration := d.movingAverageIteration(time.Now().Add(-execDuration), false, false)
	assert.InDelta(t, float64(1000), avgExecDuration, 1)

	execDurations := []time.Duration{
		time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500,
		time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500,
		time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500,
	}

	current := float64(1000)

	// Execute at 500ms for 12 requests.
	for _, execDuration = range execDurations {
		_, avgExecDuration = d.movingAverageIteration(time.Now().Add(-execDuration), false, true)
		assert.InDelta(t, current, avgExecDuration, 1)

		// Should not dip below 500, and should decrease in value by 50 each iteration.
		if current > 500 {
			current -= 50
		}
	}
}

func TestTimingAttackDelayCalculations(t *testing.T) {
	execDuration := 500 * time.Millisecond
	avgExecDurationMs := 1000.0
	expectedMinimumDelayMs := avgExecDurationMs - float64(execDuration.Milliseconds())

	ctx := &AutheliaCtx{
		Logger: logrus.NewEntry(logging.Logger()),
		Providers: Providers{
			Random: random.New(),
			Clock:  clock.New(),
		},
	}

	for i := 0; i < 100; i++ {
		delay := calculateActualDelay(ctx, execDuration, avgExecDurationMs, 250, 85, true, false)
		assert.True(t, delay >= expectedMinimumDelayMs)
		assert.True(t, delay <= expectedMinimumDelayMs+float64(85))
	}

	execDuration = 5 * time.Millisecond
	avgExecDurationMs = 5.0
	expectedMinimumDelayMs = 250 - float64(execDuration.Milliseconds())

	for i := 0; i < 100; i++ {
		delay := calculateActualDelay(ctx, execDuration, avgExecDurationMs, 250, 85, true, false)
		assert.True(t, delay >= expectedMinimumDelayMs)
		assert.True(t, delay <= expectedMinimumDelayMs+float64(250))
	}
}
