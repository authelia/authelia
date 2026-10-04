// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events

type sensitiveSliceElement struct {
	Token string `json:"token,omitempty" sensitive:"true"`
}

type sensitiveMapElement struct {
	Token string `json:"token,omitempty" sensitive:"true"`
}
