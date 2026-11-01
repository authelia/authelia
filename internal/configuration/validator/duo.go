// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateDuo validates and updates the Duo configuration.
func ValidateDuo(config *schema.Configuration, validator *schema.StructValidator) {
	if config.DuoAPI.Disable {
		return
	}

	if config.DuoAPI.Hostname == "" && config.DuoAPI.IntegrationKey == "" && config.DuoAPI.SecretKey == "" {
		config.DuoAPI.Disable = true
	}

	if config.DuoAPI.Disable {
		return
	}

	if config.DuoAPI.Hostname == "" {
		validator.Push(fmt.Errorf(errFmtDuoMissingOption, "hostname"))
	}

	if config.DuoAPI.IntegrationKey == "" {
		validator.Push(fmt.Errorf(errFmtDuoMissingOption, "integration_key"))
	}

	if config.DuoAPI.SecretKey == "" {
		validator.Push(fmt.Errorf(errFmtDuoMissingOption, "secret_key"))
	}

	if config.DuoAPI.DisableCAPinning {
		if isDuoHostname(config.DuoAPI.Hostname) {
			validator.Push(fmt.Errorf(errFmtDuoDisableCAPinningDuoHostname, config.DuoAPI.Hostname))
		} else {
			validator.PushWarning(errors.New(errDuoDisableCAPinningWarning))
		}
	}
}

// duoDomains are the domains Duo serves its API from; Duo's CA pinning must never be disabled for these.
var duoDomains = []string{"duosecurity.com", "duofederal.com"}

func isDuoHostname(hostname string) bool {
	hostname = strings.TrimSpace(hostname)

	// The Duo client uses the configured value verbatim as the URL host, so a port or userinfo must not
	// let a Duo hostname slip past the domain match.
	if u, err := url.Parse("https://" + hostname); err == nil {
		hostname = u.Hostname()
	}

	hostname = strings.ToLower(strings.TrimSuffix(hostname, "."))

	for _, domain := range duoDomains {
		if hostname == domain || strings.HasSuffix(hostname, "."+domain) {
			return true
		}
	}

	return false
}
