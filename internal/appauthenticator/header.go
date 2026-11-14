// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package appauthenticator

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

var (
	reAuthorization = regexp.MustCompile(`^AutheliaApp device_id="([A-Za-z0-9_-]{1,256})", ts="(\d{1,12})", nonce="([^"]*)", sig="([A-Za-z0-9_-]+)"$`)
	reNonce         = regexp.MustCompile(`^[A-Za-z0-9_-]{22}$`)
)

// Authorization is a parsed Authorization header.
type Authorization struct {
	DeviceID  string
	Timestamp string
	Nonce     string
	Signature string
	Time      time.Time
}

// ParseAuthorization parses an Authorization header value.
func ParseAuthorization(value []byte) (*Authorization, error) {
	m := reAuthorization.FindSubmatch(value)
	if m == nil {
		return nil, errors.New("authorization header is malformed")
	}

	if !reNonce.Match(m[3]) {
		return nil, errors.New("authorization header nonce is malformed")
	}

	ts, err := strconv.ParseInt(string(m[2]), 10, 64)
	if err != nil {
		return nil, errors.New("authorization header timestamp is malformed")
	}

	return &Authorization{
		DeviceID:  string(m[1]),
		Timestamp: string(m[2]),
		Nonce:     string(m[3]),
		Signature: string(m[4]),
		Time:      time.Unix(ts, 0).UTC(),
	}, nil
}
