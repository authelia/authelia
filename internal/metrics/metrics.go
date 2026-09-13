// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/authelia/authelia/v4/internal/regulation"
)

// Provider implementation.
type Provider interface {
	Recorder
	regulation.MetricsRecorder

	GetRegisterer() prometheus.Registerer
	GetGatherer() prometheus.Gatherer
}

// Recorder of metrics.
type Recorder interface {
	RecordRequest(statusCode, requestMethod string, elapsed time.Duration)
	RecordRequestOpenIDConnect(endpoint, statusCode string, elapsed time.Duration)
	RecordAuthz(statusCode string)
	RecordAuthenticationDuration(success bool, elapsed time.Duration)
	RecordRecoveryCodesGenerated(count int)
}

// RecoveryCodeCounts is the minimal storage-side surface the metrics provider needs to expose scrape-time gauges
// for fleet-wide recovery code adoption. The storage.Provider type satisfies this interface.
type RecoveryCodeCounts interface {
	CountUsersWithRecoveryCodes(ctx context.Context) (int, error)
	CountUsersWithLowRecoveryCodes(ctx context.Context) (int, error)
	CountUsersWithDepletedRecoveryCodes(ctx context.Context) (int, error)
}
