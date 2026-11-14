// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPrometheus(t *testing.T) {
	p, err := NewPrometheus()
	assert.NoError(t, err)
	require.NotNil(t, p)

	p.RecordRequest("400", "GET", time.Second)
	p.RecordAuthz("400")
	p.RecordAuthn(true, false, "WebAuthn")
	p.RecordAuthn(true, false, "1fa")
	p.RecordAuthn(true, false, "passkey")
	p.RecordAuthenticationDuration(true, time.Second)
	p.RecordRequestOpenIDConnect("example", "200", time.Second)
}

func TestPrometheusRecordRecoveryCodesGenerated(t *testing.T) {
	testCases := []struct {
		name     string
		counts   []int
		expected float64
	}{
		{"ShouldIgnoreZero", []int{0}, 0},
		{"ShouldIgnoreNegative", []int{-4}, 0},
		{"ShouldAddPositive", []int{10}, 10},
		{"ShouldAccumulate", []int{10, 0, 3, -1}, 13},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewPrometheus()
			require.NoError(t, err)

			for _, count := range tc.counts {
				p.RecordRecoveryCodesGenerated(count)
			}

			kind, value, found := gatherMetricValue(t, p, "authelia_recovery_codes_generated_total")
			require.True(t, found)

			assert.Equal(t, "COUNTER", kind)
			assert.Equal(t, tc.expected, value)
		})
	}
}

func TestPrometheusRecordRecoveryCodesGeneratedShouldNotPanicWithoutCounter(t *testing.T) {
	p := &Prometheus{}

	assert.NotPanics(t, func() {
		p.RecordRecoveryCodesGenerated(5)
	})
}

func TestPrometheusRegisterRecoveryCodeGauges(t *testing.T) {
	testCases := []struct {
		name     string
		provider *testRecoveryCodeCounts
		expected map[string]float64
		absent   []string
	}{
		{
			"ShouldExposeAllGauges",
			&testRecoveryCodeCounts{users: 7, low: 2, depleted: 3},
			map[string]float64{
				"authelia_recovery_codes_users_with_codes": 7,
				"authelia_recovery_codes_users_low":        2,
				"authelia_recovery_codes_users_depleted":   3,
			},
			nil,
		},
		{
			"ShouldOmitUsersGaugeOnError",
			&testRecoveryCodeCounts{users: 7, low: 2, depleted: 3, errUsers: errors.New("users failed")},
			map[string]float64{
				"authelia_recovery_codes_users_low":      2,
				"authelia_recovery_codes_users_depleted": 3,
			},
			[]string{"authelia_recovery_codes_users_with_codes"},
		},
		{
			"ShouldOmitLowGaugeOnError",
			&testRecoveryCodeCounts{users: 7, low: 2, depleted: 3, errLow: errors.New("low failed")},
			map[string]float64{
				"authelia_recovery_codes_users_with_codes": 7,
				"authelia_recovery_codes_users_depleted":   3,
			},
			[]string{"authelia_recovery_codes_users_low"},
		},
		{
			"ShouldOmitDepletedGaugeOnError",
			&testRecoveryCodeCounts{users: 7, low: 2, depleted: 3, errDepleted: errors.New("depleted failed")},
			map[string]float64{
				"authelia_recovery_codes_users_with_codes": 7,
				"authelia_recovery_codes_users_low":        2,
			},
			[]string{"authelia_recovery_codes_users_depleted"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewPrometheus()
			require.NoError(t, err)

			p.RegisterRecoveryCodeGauges(tc.provider)

			for name, value := range tc.expected {
				kind, actual, found := gatherMetricValue(t, p, name)
				require.True(t, found, name)

				assert.Equal(t, "GAUGE", kind, name)
				assert.Equal(t, value, actual, name)
			}

			for _, name := range tc.absent {
				_, _, found := gatherMetricValue(t, p, name)
				assert.False(t, found, name)
			}
		})
	}
}

func TestPrometheusRegisterRecoveryCodeGaugesShouldIgnoreNilProvider(t *testing.T) {
	p, err := NewPrometheus()
	require.NoError(t, err)

	p.RegisterRecoveryCodeGauges(nil)

	assert.Nil(t, p.recoveryCodesGaugeProvider)

	_, _, found := gatherMetricValue(t, p, "authelia_recovery_codes_users_with_codes")
	assert.False(t, found)

	p.RegisterRecoveryCodeGauges(&testRecoveryCodeCounts{})

	_, _, found = gatherMetricValue(t, p, "authelia_recovery_codes_users_with_codes")
	assert.True(t, found)
}

func TestRecoveryCodeGaugeCollectorDescribe(t *testing.T) {
	p, err := NewPrometheus()
	require.NoError(t, err)

	collector := &recoveryCodeGaugeCollector{prom: p}

	ch := make(chan *prometheus.Desc, 3)

	collector.Describe(ch)
	close(ch)

	var descs []*prometheus.Desc

	for desc := range ch {
		descs = append(descs, desc)
	}

	assert.Equal(t, []*prometheus.Desc{p.recoveryCodesUsers, p.recoveryCodesUsersLow, p.recoveryCodesUsersDepleted}, descs)
}

type testRecoveryCodeCounts struct {
	users, low, depleted          int
	errUsers, errLow, errDepleted error
}

func (c *testRecoveryCodeCounts) CountUsersWithRecoveryCodes(_ context.Context) (int, error) {
	return c.users, c.errUsers
}

func (c *testRecoveryCodeCounts) CountUsersWithLowRecoveryCodes(_ context.Context) (int, error) {
	return c.low, c.errLow
}

func (c *testRecoveryCodeCounts) CountUsersWithDepletedRecoveryCodes(_ context.Context) (int, error) {
	return c.depleted, c.errDepleted
}

func gatherMetricValue(t *testing.T, p *Prometheus, name string) (kind string, value float64, found bool) {
	t.Helper()

	families, err := p.GetGatherer().Gather()
	require.NoError(t, err)

	for _, family := range families {
		if family.GetName() != name {
			continue
		}

		metrics := family.GetMetric()
		require.Len(t, metrics, 1)

		kind = family.GetType().String()

		switch kind {
		case "COUNTER":
			return kind, metrics[0].GetCounter().GetValue(), true
		default:
			return kind, metrics[0].GetGauge().GetValue(), true
		}
	}

	return "", 0, false
}
