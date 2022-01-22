package templates

import (
	"text/template"
)

// PlainTextEmailTemplateStep1 the template of email that the user will receive for identity verification.
var PlainTextEmailTemplateStep1 *template.Template

func init() {
	t, err := template.New("text_email_template").Parse(emailPlainTextContentStep1)
	if err != nil {
		panic(err)
	}

	PlainTextEmailTemplateStep1 = t
}

const emailPlainTextContentStep1 = `
This email has been sent to you in order to validate your identity.
If you did not initiate the process your credentials might have been compromised. You should reset your password and contact an administrator.

To setup your 2FA please visit the following URL: {{.url}}

Please contact an administrator if you did not initiate the process.
`
