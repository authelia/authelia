package ntp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShouldDetectNTPOffsetTooLarge(t *testing.T) {
	assert.True(t, ntpIsOffsetTooLarge(time.Second, time.Now(), time.Now().Add(time.Second*2)))
	assert.False(t, ntpIsOffsetTooLarge(time.Second, time.Now(), time.Now()))
}
