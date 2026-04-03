package regulation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

func TestExpires(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	since := now.Add(-30 * time.Second)

	cfg := schema.Regulation{
		MaxRetries: 3,
		BanTime:    10 * time.Minute,
		FindTime:   30 * time.Second,
	}

	testCases := []struct {
		name     string
		config   schema.Regulation
		since    time.Time
		records  []model.RegulationRecord
		expected *time.Time
	}{
		{
			name:   "ShouldReturnNilWhenFailuresBelowThreshold",
			config: cfg,
			since:  since,
			records: []model.RegulationRecord{
				{Successful: false, Time: now.Add(-5 * time.Second)},
				{Successful: false, Time: now.Add(-10 * time.Second)},
			},
			expected: nil,
		},
		{
			name:   "ShouldIgnoreRecordsBeforeSince",
			config: cfg,
			since:  since,
			records: []model.RegulationRecord{
				{Successful: false, Time: now.Add(-40 * time.Second)},
				{Successful: false, Time: now.Add(-5 * time.Second)},
				{Successful: false, Time: now.Add(-10 * time.Second)},
			},
			expected: nil,
		},
		{
			name:   "ShouldExpireAtFirstFailureInWindow",
			config: cfg,
			since:  since,
			records: []model.RegulationRecord{
				{Successful: false, Time: now.Add(-2 * time.Second)},
				{Successful: false, Time: now.Add(-6 * time.Second)},
				{Successful: false, Time: now.Add(-10 * time.Second)},
			},
			expected: func() *time.Time {
				exp := now.Add(-2 * time.Second).Add(cfg.BanTime)
				return &exp
			}(),
		},
		{
			name:   "ShouldStopCountingAtSuccessfulAttempt",
			config: cfg,
			since:  since,
			records: []model.RegulationRecord{
				{Successful: false, Time: now.Add(-2 * time.Second)},
				{Successful: true, Time: now.Add(-4 * time.Second)},
				{Successful: false, Time: now.Add(-6 * time.Second)},
				{Successful: false, Time: now.Add(-8 * time.Second)},
			},
			expected: nil,
		},
		{
			name:   "ShouldIgnoreExcessFailuresBeyondMaxRetries",
			config: cfg,
			since:  since,
			records: []model.RegulationRecord{
				{Successful: false, Time: now.Add(-1 * time.Second)},
				{Successful: false, Time: now.Add(-3 * time.Second)},
				{Successful: false, Time: now.Add(-5 * time.Second)},
				{Successful: false, Time: now.Add(-7 * time.Second)},
				{Successful: false, Time: now.Add(-9 * time.Second)},
			},
			expected: func() *time.Time {
				exp := now.Add(-1 * time.Second).Add(cfg.BanTime)
				return &exp
			}(),
		},
		{
			name:     "ShouldReturnNilWhenNoRecords",
			config:   cfg,
			since:    since,
			records:  []model.RegulationRecord{},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &Regulator{
				config: tc.config,
			}

			actual := r.expires(tc.since, tc.records)

			if tc.expected == nil {
				assert.Nil(t, actual)
			} else {
				require.NotNil(t, actual)
				assert.Equal(t, *tc.expected, *actual)
			}
		})
	}
}
