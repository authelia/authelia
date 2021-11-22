package notification

import (
	"regexp"
)

const fileNotifierMode = 0600
const rfc5322DateTimeLayout = "Mon, 2 Jan 2006 15:04:05 -0700"

var (
	// RegexpValidEmail matches if the email conforms to a known standard.
	RegexpValidEmail = regexp.MustCompile(`^[0-9a-zA-Z+-_~]+@([a-zA-Z0-9]([a-zA-Z0-9-]+[a-zA-Z0-9])?\.){1,127}[a-z-A-Z]+$`)
)
