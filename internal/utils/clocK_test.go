package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTestingClock(t *testing.T) {
	c := &TestingClock{
		now: time.Unix(0, 0),
	}

	assert.Equal(t, int64(0), c.Now().Unix())
	c.now = time.Unix(20, 0)

	assert.Equal(t, int64(20), c.Now().Unix())
	assert.Equal(t, int64(20000000000), c.Now().UnixNano())

	c.Set(time.Unix(16000000, 0))

	assert.Equal(t, int64(16000000), c.Now().Unix())

	before := c.Now()

	<-c.After(time.Millisecond * 100)

	assert.Equal(t, before, c.Now())
}

func TestRealClock(t *testing.T) {
	c := &RealClock{}

	assert.WithinDuration(t, time.Now(), c.Now(), time.Second)

	before := c.Now()

	<-c.After(time.Millisecond * 100)

	after := c.Now()

	assert.WithinDuration(t, before, after, time.Millisecond*120)

	diff := after.Sub(before)

	assert.GreaterOrEqual(t, diff, time.Millisecond*100)
}
