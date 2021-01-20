package templates

import (
	"text/template"
)

// PlainTextEmailTemplate the template of email that the user will receive for identity verification.
var PlainTextEmailTemplate *template.Template

func init() {
	t, err := template.New("text_email_template").Parse(emailPlainTextContent)
	if err != nil {
		panic(err)
	}

	PlainTextEmailTemplate = t
}

const emailPlainTextContent = `
This email has been sent to you in order to validate your identity.
If you did not initiate the process your credentials might have been compromised. You should reset your password and contact an administrator.

To setup your 2FA please visit the following URL: {{.url}}

Please ignore this email if you did not initiate the process.
`
