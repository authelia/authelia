// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package duo

import "strings"

// Duo Methods.
const (
	// Push Method - The device is activated for Duo Push.
	Push = "push"
	// OTP Method - The device is capable of generating passcodes with the Duo Mobile app.
	OTP = "mobile_otp"
	// Phone Method - The device can receive phone calls.
	Phone = "phone"
	// SMS Method - The device can receive batches of SMS passcodes.
	SMS = "sms"
)

// PossibleMethods is the set of all possible Duo 2FA methods.
var PossibleMethods = []string{Push} // OTP, Phone, SMS.

// hostnameSuffixes are the domains Duo serves its API from. The client's pinned Duo CA bundle only
// applies to these; any other hostname is a third-party implementation of the Duo Auth API, whose
// certificate the pinned bundle would never accept.
var hostnameSuffixes = []string{".duosecurity.com", ".duofederal.com"}

// IsDuoHostname returns true if hostname is one of Duo's own API hostnames.
func IsDuoHostname(hostname string) bool {
	hostname = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(hostname), "."))

	for _, suffix := range hostnameSuffixes {
		if strings.HasSuffix(hostname, suffix) {
			return true
		}
	}

	return false
}
