package ntp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/utils"
)

func TestNtpIsOffsetTooLarge(t *testing.T) {
	maxOffset, _ := utils.ParseDurationString("1s")
	assert.True(t, ntpIsOffsetTooLarge(maxOffset, time.Now(), time.Now().Add(time.Second*2)))
	assert.True(t, ntpIsOffsetTooLarge(maxOffset, time.Now().Add(time.Second*2), time.Now()))
	assert.False(t, ntpIsOffsetTooLarge(maxOffset, time.Now(), time.Now()))
}

func TestNtpPacketToTime(t *testing.T) {
	resp := &ntpPacket{
		TxTimeSeconds:  60,
		TxTimeFraction: 0,
	}

	expected := time.Unix(int64(float64(60) - ntpEpochOffset), 0)

	ntpTime := ntpPacketToTime(resp)
	assert.Equal(t, expected, ntpTime)
}

func TestLeapVersionClientMode(t *testing.T) {
	v3Noleap := uint8(27)
	v4Noleap := uint8(43)
	v3leap := uint8(91)
	v4leap := uint8(107)

	assert.Equal(t, v3Noleap, ntpLeapVersionClientMode(false, ntpV3))
	assert.Equal(t, v4Noleap, ntpLeapVersionClientMode(false, ntpV4))
	assert.Equal(t, v3leap, ntpLeapVersionClientMode(true, ntpV3))
	assert.Equal(t, v4leap, ntpLeapVersionClientMode(true, ntpV4))
}
