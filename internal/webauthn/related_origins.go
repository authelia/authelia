// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webauthn

import (
	"net"
	"net/url"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/weppos/publicsuffix-go/publicsuffix"
)

// RelatedOriginLabeler is a protocol.RelatedOriginLabeler which derives the label from the registrable domain of the
// origin using the Public Suffix List, so origins under a multi-label public suffix such as 'example.co.uk' are counted
// by the label a client uses ('example') rather than the second last label ('co'). Hosts which are IP addresses, have
// a single label, or have no registrable domain are labeled by protocol.DefaultRelatedOriginLabeler.
func RelatedOriginLabeler(origin string) (label string, err error) {
	uri, err := url.Parse(origin)
	if err != nil {
		return protocol.DefaultRelatedOriginLabeler(origin)
	}

	host := strings.ToLower(strings.TrimSuffix(uri.Hostname(), "."))

	if host == "" || net.ParseIP(host) != nil || !strings.Contains(host, ".") {
		return protocol.DefaultRelatedOriginLabeler(origin)
	}

	domain, err := publicsuffix.Parse(host)
	if err != nil || domain == nil || domain.SLD == "" {
		return protocol.DefaultRelatedOriginLabeler(origin)
	}

	return domain.SLD, nil
}
