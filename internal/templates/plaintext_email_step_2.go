package templates

import (
	"text/template"
)

// PlainTextEmailTemplateStep2 the template of email that the user will receive for identity verification.
var PlainTextEmailTemplateStep2 *template.Template

func init() {
	t, err := template.New("text_email_template").Parse(emailPlainTextContentStep2)
	if err != nil {
		panic(err)
	}

	PlainTextEmailTemplateStep2 = t
}

const emailPlainTextContentStep2 = `
Your password has been successfully reset.
If you did not initiate the process your credentials might have been compromised. You should reset your password and contact an administrator.

Please contact an administrator if you did not initiate the process.
`
