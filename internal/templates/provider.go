package templates

import (
	"embed"
	"fmt"
	"text/template"
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

// LoadTemplatedAssets takes an embed.FS and loads each templated asset document into a Template.
func (p *Provider) LoadTemplatedAssets(fs embed.FS) (err error) {
	var (
		data []byte
	)

	if data, err = fs.ReadFile("public_html/index.html"); err != nil {
		return err
	}

	if p.templates.asset.index, err = template.
		New("assets/public_html/index.html").
		Funcs(FuncMap()).
		Parse(string(data)); err != nil {
		return err
	}

	if data, err = fs.ReadFile("public_html/api/index.html"); err != nil {
		return err
	}

	if p.templates.asset.api.index, err = template.
		New("assets/public_html/api/index.html").
		Funcs(FuncMap()).
		Parse(string(data)); err != nil {
		return err
	}

	if data, err = fs.ReadFile("public_html/api/openapi.yml"); err != nil {
		return err
	}

	if p.templates.asset.api.spec, err = template.
		New("api/public_html/openapi.yaml").
		Funcs(FuncMap()).
		Parse(string(data)); err != nil {
		return err
	}

	return nil
}

// GetAssetIndexTemplate returns a Template used to generate the React index document.
func (p *Provider) GetAssetIndexTemplate() (t Template) {
	return p.templates.asset.index
}

// GetAssetOpenAPIIndexTemplate returns a Template used to generate the OpenAPI index document.
func (p *Provider) GetAssetOpenAPIIndexTemplate() (t Template) {
	return p.templates.asset.api.index
}

// GetAssetOpenAPISpecTemplate returns a Template used to generate the OpenAPI specification document.
func (p *Provider) GetAssetOpenAPISpecTemplate() (t Template) {
	return p.templates.asset.api.spec
}

// GetEventEmailTemplate returns an EmailTemplate used for generic event notifications.
func (p *Provider) GetEventEmailTemplate() (t *EmailTemplate) {
	return p.templates.notification.event
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
