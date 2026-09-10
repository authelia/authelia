// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConformanceProgress(t *testing.T) {
	patience := time.Second * 30
	start := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

	t.Run("ShouldStallOnAnUnrecognizedPageThatNeverMoves", func(t *testing.T) {
		progress := newConformanceProgress(patience, start)

		assert.False(t, progress.observe(start, "https://stuck.example.com", false))
		assert.False(t, progress.observe(start.Add(time.Second*29), "https://stuck.example.com", false))
		assert.True(t, progress.observe(start.Add(time.Second*31), "https://stuck.example.com", false))
	})

	t.Run("ShouldNotStallWhileTheURLKeepsChanging", func(t *testing.T) {
		progress := newConformanceProgress(patience, start)

		for i, at := range []time.Duration{0, time.Second * 25, time.Second * 50, time.Second * 75} {
			assert.Falsef(t, progress.observe(start.Add(at), "https://hop.example.com/"+string(rune('a'+i)), false), "hop %d", i)
		}
	})

	t.Run("ShouldNotStallOnAPageTheDriverIsActingOn", func(t *testing.T) {
		progress := newConformanceProgress(patience, start)

		assert.False(t, progress.observe(start, "https://login.example.com/consent", true))
		assert.False(t, progress.observe(start.Add(time.Second*45), "https://login.example.com/consent", true))
		assert.False(t, progress.observe(start.Add(time.Second*60), "https://login.example.com/consent", true))
	})

	t.Run("ShouldStallOnceAPageStopsBeingActedOn", func(t *testing.T) {
		progress := newConformanceProgress(patience, start)

		assert.False(t, progress.observe(start, "https://login.example.com/consent", true))
		assert.False(t, progress.observe(start.Add(time.Second*10), "https://login.example.com/consent", false))
		assert.True(t, progress.observe(start.Add(time.Second*41), "https://login.example.com/consent", false))
	})
}
