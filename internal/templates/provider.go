package templates

import (
	"fmt"
	th "html/template"
	"io/fs"
	"path"
	tt "text/template"
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
func (p *Provider) LoadTemplatedAssets(fs fs.ReadFileFS) (err error) {
	if p.templates.asset.index, err = parseTemplatedAsset(fs, AssetPathIndex); err != nil {
		return err
	}

	if p.templates.asset.api.index, err = parseTemplatedAsset(fs, AssetPathAPIIndex); err != nil {
		return err
	}

	if p.templates.asset.api.spec, err = parseTemplatedAsset(fs, AssetPathAPISpec); err != nil {
		return err
	}

	return nil
}

func parseTemplatedAsset(fs fs.ReadFileFS, asset string) (t *tt.Template, err error) {
	var data []byte

	if data, err = fs.ReadFile(asset); err != nil {
		return nil, fmt.Errorf("error occurred loading template '%s%s': %w", assetPathPrefix, asset, err)
	}

	if t, err = tt.New(assetPathPrefix + asset).Funcs(FuncMap()).Parse(string(data)); err != nil {
		return nil, fmt.Errorf("error occurred loading template '%s%s': %w", assetPathPrefix, asset, err)
	}

	return t, nil
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

// GetIdentityVerificationJWTEmailTemplate returns the EmailTemplate for Identity Verification notifications.
func (p *Provider) GetIdentityVerificationJWTEmailTemplate() (t *EmailTemplate) {
	return p.templates.notification.jwtIdentityVerification
}

// GetIdentityVerificationOTCEmailTemplate returns the EmailTemplate for Identity Verification notifications.
func (p *Provider) GetIdentityVerificationOTCEmailTemplate() (t *EmailTemplate) {
	return p.templates.notification.otcIdentityVerification
}

// GetEventEmailTemplate returns an EmailTemplate used for generic event notifications.
func (p *Provider) GetEventEmailTemplate() (t *EmailTemplate) {
	return p.templates.notification.event
}

// GetOpenIDConnectAuthorizeResponseFormPostTemplate returns a Template used to generate the OpenID Connect 1.0 Form Post Authorize Response.
func (p *Provider) GetOpenIDConnectAuthorizeResponseFormPostTemplate() (t *th.Template) {
	return p.templates.oidc.formpost
}

func (p *Provider) load() (err error) {
	var errs []error

	if p.templates.notification.jwtIdentityVerification, err = loadEmailTemplate(TemplateNameEmailIdentityVerificationJWT, p.config.EmailTemplatesPath); err != nil {
		errs = append(errs, fmt.Errorf("error occurred loading '%s' email template: %w", TemplateNameEmailIdentityVerificationJWT, err))
	}

	if p.templates.notification.otcIdentityVerification, err = loadEmailTemplate(TemplateNameEmailIdentityVerificationOTC, p.config.EmailTemplatesPath); err != nil {
		errs = append(errs, fmt.Errorf("error occurred loading '%s' email template: %w", TemplateNameEmailIdentityVerificationOTC, err))
	}

	if p.templates.notification.event, err = loadEmailTemplate(TemplateNameEmailEvent, p.config.EmailTemplatesPath); err != nil {
		errs = append(errs, fmt.Errorf("error occurred loading '%s' email template: %w", TemplateNameEmailEvent, err))
	}

	var data []byte

	if data, err = embedFS.ReadFile(path.Join("embed", TemplateCategoryOpenIDConnect, TemplateNameOIDCAuthorizeFormPost)); err != nil {
		errs = append(errs, err)
	} else if p.templates.oidc.formpost, err = th.
		New("oidc/AuthorizeResponseFormPost.html").
		Funcs(FuncMap()).
		Parse(string(data)); err != nil {
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
