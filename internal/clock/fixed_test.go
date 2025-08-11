package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTestingClock(t *testing.T) {
	c := NewFixed(time.Unix(0, 0))

	assert.Equal(t, int64(0), c.Now().Unix())
	c.now = time.Unix(20, 0)

	assert.Equal(t, int64(20), c.Now().Unix())
	assert.Equal(t, int64(20000000000), c.Now().UnixNano())

	c.Set(time.Unix(16000000, 0))

	assert.Equal(t, int64(16000000), c.Now().Unix())

	before := c.Now()

	<-c.After(time.Millisecond * 100)

	assert.Equal(t, before, c.Now())

	done := make(chan struct{})

	value := false

	c.AfterFunc(time.Millisecond*20, func() {
		value = true

		close(done)
	})

	select {
	case <-done:
		t.Fatal("AfterFunc executed synchronously")
	default:
		assert.Equal(t, false, value)
	}

	select {
	case <-done:
		assert.Equal(t, true, value)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("AfterFunc didn't execute within expected time")
	}
}
