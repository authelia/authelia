// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package notification

const (
	fileNotifierMode   = 0600
	fileNotifierHeader = "Date: %s\nRecipient: %s\nSubject: %s\n"
)

const (
	smtpPortSUBMISSIONS = 465
)

const (
	posixNewLine = "\n"
)

var (
	posixDoubleNewLine = []byte(posixNewLine + posixNewLine)
)
