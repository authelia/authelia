// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"time"

	"github.com/authelia/authelia/v4/internal/regulation"
)

// Provider implementation.
type Provider interface {
	Recorder
	regulation.MetricsRecorder
}

// Recorder of metrics.
type Recorder interface {
	RecordRequest(statusCode, requestMethod string, elapsed time.Duration)
	RecordAuthz(statusCode string)
	RecordAuthenticationDuration(success bool, elapsed time.Duration)
}
