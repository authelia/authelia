// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package duo

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
