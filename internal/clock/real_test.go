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
