package templates

import (
	"fmt"
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

// GetEventEmailTemplate returns the EmailTemplate for Event notifications.
func (p *Provider) GetEventEmailTemplate() (t *EmailTemplate) {
	return p.templates.notification.event
}

// GetIdentityVerificationEmailTemplate returns the EmailTemplate for Identity Verification notifications.
func (p *Provider) GetIdentityVerificationEmailTemplate() (t *EmailTemplate) {
	return p.templates.notification.identityVerification
}

// GetOneTimePasswordEmailTemplate returns the EmailTemplate for One Time Password notifications.
func (p *Provider) GetOneTimePasswordEmailTemplate() (t *EmailTemplate) {
	return p.templates.notification.otp
}

func (p *Provider) load() (err error) {
	var errs []error

	if p.templates.notification.identityVerification, err = loadEmailTemplate(TemplateNameEmailIdentityVerification, p.config.EmailTemplatesPath); err != nil {
		errs = append(errs, err)
	}

	if p.templates.notification.otp, err = loadEmailTemplate(TemplateNameEmailOneTimePassword, p.config.EmailTemplatesPath); err != nil {
		errs = append(errs, err)
	}

	if p.templates.notification.event, err = loadEmailTemplate(TemplateNameEmailEvent, p.config.EmailTemplatesPath); err != nil {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		for i, e := range errs {
			if i == 0 {
				err = e
				continue
			}

			err = fmt.Errorf("%v, %w", err, e)
		}

		return fmt.Errorf("one or more errors occurred loading templates: %w", err)
	}

	return nil
}
