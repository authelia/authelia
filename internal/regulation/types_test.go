package regulation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBan_NewBanAndAccessors(t *testing.T) {
	ts := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

	testCases := []struct {
		name    string
		ban     BanType
		value   string
		expires *time.Time
	}{
		{
			name:    "ShouldCreateBanWithGivenValues",
			ban:     BanTypeIP,
			value:   "1.2.3.4",
			expires: &ts,
		},
		{
			name:    "ShouldCreateUserBanWithoutExpiry",
			ban:     BanTypeUser,
			value:   "alice",
			expires: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := NewBan(tc.ban, tc.value, tc.expires)
			assert.NotNil(t, b)

			assert.Equal(t, tc.ban, b.Type())
			assert.Equal(t, tc.value, b.Value())
			assert.Equal(t, tc.expires, b.Expires())
			assert.Equal(t, tc.ban != BanTypeNone, b.IsBanned())

			expected := FormatExpiresLong(tc.expires)
			assert.Equal(t, expected, b.FormatExpires())
		})
	}
}

func TestBan_NilReceiverDefaults(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "ShouldReturnDefaultsWhenNilReceiver",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var b *Ban

			assert.Equal(t, "", b.Value())
			assert.Equal(t, BanTypeNone, b.Type())
			assert.Nil(t, b.Expires())

			assert.Equal(t, FormatExpiresLong(nil), b.FormatExpires())
			assert.False(t, b.IsBanned())
		})
	}
}

func TestBan_IsBannedAndNoneType(t *testing.T) {
	testCases := []struct {
		name  string
		ban   *Ban
		isBan bool
	}{
		{
			name:  "ShouldReturnFalseWhenBanTypeNone",
			ban:   NewBan(BanTypeNone, "ignored", nil),
			isBan: false,
		},
		{
			name:  "ShouldReturnTrueWhenIPBan",
			ban:   NewBan(BanTypeIP, "203.0.113.1", nil),
			isBan: true,
		},
		{
			name:  "ShouldReturnTrueWhenUserBan",
			ban:   NewBan(BanTypeUser, "bob", nil),
			isBan: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.isBan, tc.ban.IsBanned())
		})
	}
}
