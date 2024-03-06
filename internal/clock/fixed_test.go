package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTestingClock(t *testing.T) {
	c := &Fixed{
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
