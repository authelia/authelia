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

// ExecuteEmailEnvelope writes the envelope template to the given io.Writer.
func (p Provider) ExecuteEmailEnvelope(wr io.Writer, data EmailEnvelopeValues) (err error) {
	return p.templates.notification.envelope.Execute(wr, data)
}

// ExecuteEmailPasswordResetTemplate writes the password reset template to the given io.Writer.
func (p Provider) ExecuteEmailPasswordResetTemplate(wr io.Writer, data EmailPasswordResetValues, format Format) (err error) {
	return p.templates.notification.passwordReset.Get(format).Execute(wr, data)
}

// ExecuteEmailIdentityVerificationTemplate writes the identity verification template to the given io.Writer.
func (p Provider) ExecuteEmailIdentityVerificationTemplate(wr io.Writer, data EmailIdentityVerificationValues, format Format) (err error) {
	return p.templates.notification.identityVerification.Get(format).Execute(wr, data)
}

func (p *Provider) load() (err error) {
	var errs []error

	if p.templates.notification.envelope, err = loadTemplate(TemplateNameEmailEnvelope, TemplateCategoryNotifications, p.config.EmailTemplatesPath); err != nil {
		errs = append(errs, err)
	}

	if p.templates.notification.identityVerification.txt, err = loadTemplate(TemplateNameEmailIdentityVerificationTXT, TemplateCategoryNotifications, p.config.EmailTemplatesPath); err != nil {
		errs = append(errs, err)
	}

	if p.templates.notification.identityVerification.html, err = loadTemplate(TemplateNameEmailIdentityVerificationHTML, TemplateCategoryNotifications, p.config.EmailTemplatesPath); err != nil {
		errs = append(errs, err)
	}

	if p.templates.notification.passwordReset.txt, err = loadTemplate(TemplateNameEmailPasswordResetTXT, TemplateCategoryNotifications, p.config.EmailTemplatesPath); err != nil {
		errs = append(errs, err)
	}

	if p.templates.notification.passwordReset.html, err = loadTemplate(TemplateNameEmailPasswordResetHTML, TemplateCategoryNotifications, p.config.EmailTemplatesPath); err != nil {
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
