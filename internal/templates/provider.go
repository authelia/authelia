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

// GetPasswordResetEmailTemplate returns the EmailTemplate for Password Reset notifications.
func (p *Provider) GetPasswordResetEmailTemplate() (t *EmailTemplate) {
	return p.templates.notification.passwordReset
}

// GetIdentityVerificationEmailTemplate returns the EmailTemplate for Identity Verification notifications.
func (p *Provider) GetIdentityVerificationEmailTemplate() (t *EmailTemplate) {
	return p.templates.notification.identityVerification
}

func (p *Provider) load() (err error) {
	var errs []error

	if p.templates.notification.identityVerification, err = loadEmailTemplate(TemplateNameEmailIdentityVerification, p.config.EmailTemplatesPath); err != nil {
		errs = append(errs, err)
	}

	if p.templates.notification.passwordReset, err = loadEmailTemplate(TemplateNameEmailPasswordReset, p.config.EmailTemplatesPath); err != nil {
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
