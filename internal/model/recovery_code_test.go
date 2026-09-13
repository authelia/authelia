// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/random"
)

func TestNewRecoveryCode(t *testing.T) {
	ctx := &TestContext{
		Context: context.Background(),
		ip:      net.ParseIP("127.0.0.1"),
		clock:   clock.NewFixed(time.Unix(1700000000, 0)),
		random:  random.NewMathematical(),
	}

	code, err := NewRecoveryCode(ctx, "alice")

	assert.NoError(t, err)
	assert.Equal(t, "alice", code.Username)
	assert.Equal(t, time.Unix(1700000000, 0), code.CreatedAt)
	assert.Empty(t, code.Signature)
	assert.False(t, code.ConsumedAt.Valid)
	assert.False(t, code.RevokedAt.Valid)

	parts := strings.Split(code.Plaintext, "-")
	assert.Len(t, parts, 2, "plaintext should be split into two halves by a hyphen")
	assert.Len(t, parts[0], RecoveryCodeLength/2)
	assert.Len(t, parts[1], RecoveryCodeLength/2)

	for _, c := range strings.Join(parts, "") {
		assert.True(t, strings.ContainsRune(random.CharSetUnambiguousUpper, c), "code char %q must be in the unambiguous upper alphabet", c)
	}
}

func TestRecoveryCodeConsume(t *testing.T) {
	ctx := &TestContext{
		Context: context.Background(),
		ip:      net.ParseIP("10.0.0.1"),
		clock:   clock.NewFixed(time.Unix(1700000000, 0)),
		random:  random.NewMathematical(),
	}

	code := &RecoveryCode{Username: "alice"}

	code.Consume(ctx)

	assert.True(t, code.ConsumedAt.Valid)
	assert.Equal(t, time.Unix(1700000000, 0), code.ConsumedAt.Time)
	assert.Equal(t, NewNullIP(net.ParseIP("10.0.0.1")), code.ConsumedIP)
	assert.False(t, code.RevokedAt.Valid)
}

func TestRecoveryCodeRevoke(t *testing.T) {
	ctx := &TestContext{
		Context: context.Background(),
		ip:      net.ParseIP("10.0.0.2"),
		clock:   clock.NewFixed(time.Unix(1700000000, 0)),
		random:  random.NewMathematical(),
	}

	code := &RecoveryCode{Username: "alice"}

	code.Revoke(ctx)

	assert.True(t, code.RevokedAt.Valid)
	assert.Equal(t, time.Unix(1700000000, 0), code.RevokedAt.Time)
	assert.Equal(t, NewNullIP(net.ParseIP("10.0.0.2")), code.RevokedIP)
	assert.False(t, code.ConsumedAt.Valid)
}

func TestNormalizeRecoveryCode(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"abcde-fghij", "ABCDEFGHIJ"},
		{"ABCDE-FGHIJ", "ABCDEFGHIJ"},
		{"abcdefghij", "ABCDEFGHIJ"},
		{"ABCDE_FGHIJ", "ABCDEFGHIJ"},
		{" ABCDE FGHIJ ", "ABCDEFGHIJ"},
		{"abcde\tFGHIJ", "ABCDEFGHIJ"},
		{"abcde\n-fghij", "ABCDEFGHIJ"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, NormalizeRecoveryCode(tt.input))
		})
	}
}
