package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRealClock(t *testing.T) {
	c := New()

	assert.WithinDuration(t, time.Now(), c.Now(), time.Second)

	before := c.Now()

	<-c.After(time.Millisecond * 100)

	after := c.Now()

	assert.WithinDuration(t, before, after, time.Millisecond*120)

	diff := after.Sub(before)

	assert.GreaterOrEqual(t, diff, time.Millisecond*100)
}
