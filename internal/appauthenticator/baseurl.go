// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package appauthenticator

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	reSegment = regexp.MustCompile(`^[A-Za-z0-9._~-]+$`)
	reDNSName = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9.-]*[A-Za-z0-9])?$`)
	reIPv6    = regexp.MustCompile(`^[0-9a-fA-F:.]+$`)
	rePort    = regexp.MustCompile(`^[1-9][0-9]{0,4}$`)
)

// NewBaseURL derives the normalized base URL from the portal URL.
func NewBaseURL(portalURL *url.URL) (string, error) {
	if portalURL == nil {
		return "", errors.New("portal url is nil")
	}

	scheme := portalURL.Scheme
	if scheme != "https" {
		return "", fmt.Errorf("portal url scheme must be https but was '%s'", scheme)
	}

	host, port, err := normalizeAuthority(portalURL)
	if err != nil {
		return "", err
	}

	var b strings.Builder

	b.WriteString(scheme)
	b.WriteString("://")
	b.WriteString(host)

	if port != "" {
		b.WriteString(":")
		b.WriteString(port)
	}

	p := portalURL.EscapedPath()

	p = strings.TrimSuffix(p, "/")

	if p != "" {
		if !strings.HasPrefix(p, "/") {
			return "", errors.New("portal url path is invalid")
		}

		for _, seg := range strings.Split(p[1:], "/") {
			if seg == "." || seg == ".." || !reSegment.MatchString(seg) {
				return "", fmt.Errorf("portal url path segment '%s' is not permitted", seg)
			}

			b.WriteString("/")
			b.WriteString(seg)
		}
	}

	return b.String(), nil
}

// QRText builds the QR text for the base URL and registration random.
func QRText(baseURL, random string) string {
	return qrScheme + strings.TrimPrefix(baseURL, "https://") + qrPath + random
}

func normalizeAuthority(portalURL *url.URL) (host, port string, err error) {
	host = strings.ToLower(portalURL.Hostname())

	switch {
	case strings.Contains(host, ":"):
		if !reIPv6.MatchString(host) {
			return "", "", fmt.Errorf("portal url host '%s' is not permitted", host)
		}

		host = "[" + host + "]"
	case !reDNSName.MatchString(host):
		return "", "", fmt.Errorf("portal url host '%s' is not permitted", host)
	}

	port = portalURL.Port()

	if port == "" {
		return host, "", nil
	}

	n, _ := strconv.Atoi(port)

	if !rePort.MatchString(port) || n > 65535 {
		return "", "", fmt.Errorf("portal url port '%s' is not permitted", port)
	}

	if n == 443 {
		port = ""
	}

	return host, port, nil
}
