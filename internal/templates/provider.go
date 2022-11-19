package templates

import (
	"fmt"
	"io"
)

// New creates a new templates' provider.
func New(config Config) (provider *Provider, err error) {
	provider = &Provider{
		config: config,
	}

	if err = provider.load(); err != nil {
		return nil, err
	}

	return provider, nil
}

// Provider of templates.
type Provider struct {
	config    Config
	templates Templates
}

func (p *Provider) GetEmailPasswordTemplate() (t *EmailTemplate) {
	return p.templates.notification.passwordReset
}

func (p *Provider) GetEmailIdentityVerificationTemplate() (t *EmailTemplate) {
	return p.templates.notification.identityVerification
}

// ExecuteEmailPasswordResetTemplate writes the password reset template to the given io.Writer.
func (p *Provider) ExecuteEmailPasswordResetTemplate(wr io.Writer, data EmailPasswordResetValues, format Format) (err error) {
	return p.templates.notification.passwordReset.Get(format).Execute(wr, data)
}

// ExecuteEmailIdentityVerificationTemplate writes the identity verification template to the given io.Writer.
func (p *Provider) ExecuteEmailIdentityVerificationTemplate(wr io.Writer, data EmailIdentityVerificationValues, format Format) (err error) {
	return p.templates.notification.identityVerification.Get(format).Execute(wr, data)
}

func (p *Provider) load() (err error) {
	var errs []error

	if p.templates.notification.identityVerification, err = loadEmailTemplate(TemplateNameEmailIdentityVerification, p.config.EmailTemplatesPath); err != nil {
		errs = append(errs, err)
	}

	if p.templates.notification.passwordReset, err = loadEmailTemplate(TemplateNameEmailPasswordReset, p.config.EmailTemplatesPath); err != nil {
		errs = append(errs, err)
	}

	if len(errs) == 0 {
		return nil
	}

	for i, e := range errs {
		if i == 0 {
			err = e
			continue
		}

		err = fmt.Errorf("%v, %w", err, e)
	}

	return fmt.Errorf("one or more errors occurred loading templates: %w", err)
}
