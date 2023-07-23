package middlewares

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimingAttackDelayAverages(t *testing.T) {
	delayer := NewTimingAttackDelayer("test", time.Second, time.Millisecond*250, time.Millisecond*85, 10)

	expected := float64(1000)

	elapsedDurations := []time.Duration{
		time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500,
		time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500,
		time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500,
		time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500,
		time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500,
	}

	// Execute at 500ms.
	for i, have := range elapsedDurations {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if i == 0 {
				assert.Equal(t, expected, delayer.avg(have, false))
			} else {
				assert.Equal(t, expected, delayer.avg(have, true))

				// Should not dip below 500, and should decrease in value by 50 each iteration where it was successful.
				if expected > 500 {
					expected -= 50
				}
			}
		})
	}
}

func TestTimingAttackDelayCalculations(t *testing.T) {
	min := time.Millisecond * 250
	max := time.Millisecond * 85
	avg := time.Second

	delayer := NewTimingAttackDelayer("test", avg, min, max, 10)
	elapsed := 500 * time.Millisecond

	expectedMin := avg - elapsed

	for i := 0; i < 100; i++ {
		delay := delayer.actual(elapsed, delayer.avg(elapsed, false), false)
		assert.GreaterOrEqual(t, delay, expectedMin)
		assert.LessOrEqual(t, delay, expectedMin+max)
	}

	elapsed = time.Millisecond * 5
	avg = time.Millisecond * 5

	expectedMin = min - elapsed

	delayer = NewTimingAttackDelayer("test", avg, min, max, 10)

	for i := 0; i < 100; i++ {
		delay := delayer.actual(elapsed, delayer.avg(elapsed, false), false)
		assert.GreaterOrEqual(t, delay, expectedMin)
		assert.LessOrEqual(t, delay, expectedMin+max)
	}
}

/*

 */
