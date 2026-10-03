// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"strconv"
	"strings"

	"github.com/authelia/authelia/v4/internal/events"
)

// Sign returns the value of the X-Authelia-Signature header for a payload, or an empty string when no secret is
// configured. The signed input is the timestamp, a full stop, then the raw body bytes exactly as transmitted. The
// timestamp is part of the signed input so a captured request cannot be replayed indefinitely. The v1 prefix versions
// the scheme so an additional construction can be added later without breaking receivers.
func Sign(secret, algorithm string, timestamp int64, body []byte) (signature string) {
	if secret == "" {
		return ""
	}

	var mac hash.Hash

	switch algorithm {
	case events.SignatureAlgorithmSHA512:
		mac = hmac.New(sha512.New, []byte(secret))
	default:
		mac = hmac.New(sha256.New, []byte(secret))
	}

	ts := strconv.FormatInt(timestamp, 10)

	mac.Write([]byte(ts))
	mac.Write([]byte("."))
	mac.Write(body)

	builder := &strings.Builder{}

	builder.WriteString("t=")
	builder.WriteString(ts)
	builder.WriteString(",v1=")
	builder.WriteString(hex.EncodeToString(mac.Sum(nil)))

	return builder.String()
}
