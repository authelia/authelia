package ntp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/utils"
)

func TestShould(t *testing.T) {
	maxOffset, _ := utils.ParseDurationString("1s")
	assert.True(t, ntpIsOffsetTooLarge(maxOffset, time.Now(), time.Now().Add(time.Second*2)))
	assert.False(t, ntpIsOffsetTooLarge(maxOffset, time.Now(), time.Now()))
}
