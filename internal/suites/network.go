// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

const networkFailureLimit = 200

type networkRequest struct {
	Method string
	URL    string
}

type networkFailure struct {
	Time     time.Time                     `json:"time"`
	Method   string                        `json:"method"`
	URL      string                        `json:"url"`
	Type     proto.NetworkResourceType     `json:"type"`
	Error    string                        `json:"error"`
	Canceled bool                          `json:"canceled,omitempty"`
	Blocked  proto.NetworkBlockedReason    `json:"blocked,omitempty"`
	CORS     *proto.NetworkCorsErrorStatus `json:"cors,omitempty"`
}

type networkRecorder struct {
	mutex    sync.Mutex
	pending  map[proto.NetworkRequestID]networkRequest
	failures []networkFailure
}

var networkRecorders sync.Map

func recordNetworkFailures(page *rod.Page) {
	recorder := &networkRecorder{pending: map[proto.NetworkRequestID]networkRequest{}, failures: []networkFailure{}}

	networkRecorders.Store(page.TargetID, recorder)

	wait := page.EachEvent(
		func(e *proto.NetworkRequestWillBeSent) {
			recorder.mutex.Lock()
			defer recorder.mutex.Unlock()

			recorder.pending[e.RequestID] = networkRequest{Method: e.Request.Method, URL: e.Request.URL}
		},
		func(e *proto.NetworkLoadingFinished) {
			recorder.mutex.Lock()
			defer recorder.mutex.Unlock()

			delete(recorder.pending, e.RequestID)
		},
		func(e *proto.NetworkLoadingFailed) {
			recorder.mutex.Lock()
			defer recorder.mutex.Unlock()

			request := recorder.pending[e.RequestID]

			delete(recorder.pending, e.RequestID)

			if len(recorder.failures) >= networkFailureLimit {
				return
			}

			recorder.failures = append(recorder.failures, networkFailure{
				Time:     time.Now(),
				Method:   request.Method,
				URL:      request.URL,
				Type:     e.Type,
				Error:    e.ErrorText,
				Canceled: e.Canceled,
				Blocked:  e.BlockedReason,
				CORS:     e.CorsErrorStatus,
			})
		},
		func(e *proto.InspectorDetached) bool {
			return true
		},
	)

	go func() {
		wait()

		networkRecorders.Delete(page.TargetID)
	}()
}

func networkFailures(page *rod.Page) ([]byte, bool) {
	value, ok := networkRecorders.Load(page.TargetID)
	if !ok {
		return nil, false
	}

	recorder := value.(*networkRecorder) //nolint:forcetypeassert

	recorder.mutex.Lock()
	defer recorder.mutex.Unlock()

	data, err := json.MarshalIndent(recorder.failures, "", "  ")
	if err != nil {
		return nil, false
	}

	return data, true
}
