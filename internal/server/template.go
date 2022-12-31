package server

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/valyala/fasthttp"
	"golang.org/x/exp/maps"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/utils"
)

// Template represents Go template to parse and serve html files.
type Template struct {
	*template.Template

	opts             *TemplateOptions
	fileExt          string
	isDevEnvironment bool
	isAssetsSwagger  bool
}

// ParseFile parses file into to the template with the specific path.
func (tmpl *Template) ParseFile(path string) *Template {
	logger := logging.Logger()

	a, err := assets.Open(path)
	if err != nil {
		logger.Fatalf("Unable to open %s: %s", path, err)
	}

	b, err := io.ReadAll(a)
	if err != nil {
		logger.Fatalf("Unable to read %s: %s", path, err)
	}

	tmpl.fileExt = filepath.Ext(path)
	tmpl.isAssetsSwagger = strings.HasPrefix(path, assetsSwagger)

	if tmpl.opts.AssetPath != "" {
		if _, err = os.Stat(filepath.Join(tmpl.opts.AssetPath, fileLogo)); err == nil {
			tmpl.opts.LogoOverride = t
		}
	}

	tmpl.Template, err = template.New("file").Parse(string(b))
	if err != nil {
		logger.Fatalf("Unable to parse %s template: %s", path, err)
	}

	return tmpl
}

// Handler passes information between the backend and frontend
// and generate a nonce to support a restrictive CSP while using material-ui.
func (tmpl *Template) Handler(data ...map[string]any) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		switch tmpl.fileExt {
		case extHTML:
			ctx.SetContentTypeTextHTML()
		case extJSON:
			ctx.SetContentTypeApplicationJSON()
		default:
			ctx.SetContentTypeTextPlain()
		}

		nonce := utils.RandomString(32, utils.CharSetAlphaNumeric)

		switch {
		case tmpl.isAssetsSwagger:
			ctx.Response.Header.Set(fasthttp.HeaderContentSecurityPolicy, fmt.Sprintf(tmplCSPSwagger, nonce, nonce))
		case ctx.Configuration.Server.Headers.CSPTemplate != "":
			ctx.Response.Header.Set(fasthttp.HeaderContentSecurityPolicy, strings.ReplaceAll(ctx.Configuration.Server.Headers.CSPTemplate, placeholderCSPNonce, nonce))
		case tmpl.isDevEnvironment:
			ctx.Response.Header.Set(fasthttp.HeaderContentSecurityPolicy, fmt.Sprintf(tmplCSPDevelopment, nonce))
		default:
			ctx.Response.Header.Set(fasthttp.HeaderContentSecurityPolicy, fmt.Sprintf(tmplCSPDefault, nonce))
		}

		commonData := tmpl.opts.CommonData(ctx.BasePath(), ctx.RootURLSlash().String(), nonce).toMap()

		mergedData := make(map[string]any)
		for _, d := range append(data, commonData) {
			maps.Copy(mergedData, d)
		}

		if err := tmpl.Execute(ctx.Response.BodyWriter(), mergedData); err != nil {
			ctx.RequestCtx.Error("an error occurred", 503)
			ctx.Logger.Errorf("Unable to execute template: %v", err)

			return
		}
	}
}

// NewTemplate returns a new Template instance.
func NewTemplate(config *schema.Configuration) *Template {
	return &Template{
		opts:             NewTemplateOptions(config),
		isDevEnvironment: os.Getenv(environment) == dev,
	}
}

func writeHealthCheckEnv(disabled bool, scheme, host, path string, port int) (err error) {
	if disabled {
		return nil
	}

	_, err = os.Stat("/app/healthcheck.sh")
	if err != nil {
		return nil
	}

	_, err = os.Stat("/app/.healthcheck.env")
	if err != nil {
		return nil
	}

	file, err := os.OpenFile("/app/.healthcheck.env", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	if host == "0.0.0.0" {
		host = localhost
	} else if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}

	_, err = file.WriteString(fmt.Sprintf(healthCheckEnv, scheme, host, port, path))

	return err
}

// NewTemplateOptions returns a new *TemplateOptions.
func NewTemplateOptions(config *schema.Configuration) (opts *TemplateOptions) {
	opts = &TemplateOptions{
		AssetPath:              config.Server.AssetPath,
		DuoSelfEnrollment:      f,
		LogoOverride:           f,
		RememberMe:             strconv.FormatBool(config.Session.RememberMeDuration != schema.RememberMeDisabled),
		ResetPassword:          strconv.FormatBool(!config.AuthenticationBackend.PasswordReset.Disable),
		ResetPasswordCustomURL: config.AuthenticationBackend.PasswordReset.CustomURL.String(),
		Theme:                  config.Theme,
	}

	if !config.DuoAPI.Disable {
		opts.DuoSelfEnrollment = strconv.FormatBool(config.DuoAPI.EnableSelfEnrollment)
	}

	return opts
}

// TemplateOptions is a struct which is used for many templated files.
type TemplateOptions struct {
	AssetPath              string
	DuoSelfEnrollment      string
	RememberMe             string
	ResetPassword          string
	ResetPasswordCustomURL string
	Session                string
	Theme                  string
	LogoOverride           string
}

// CommonData returns a TemplateCommonData with the dynamic options.
func (options *TemplateOptions) CommonData(base, baseURL, nonce string) *TemplateCommonData {
	return &TemplateCommonData{
		Base:                   base,
		BaseURL:                baseURL,
		CSPNonce:               nonce,
		AssetPath:              options.AssetPath,
		DuoSelfEnrollment:      options.DuoSelfEnrollment,
		RememberMe:             options.RememberMe,
		ResetPassword:          options.ResetPassword,
		ResetPasswordCustomURL: options.ResetPasswordCustomURL,
		Session:                options.Session,
		Theme:                  options.Theme,
	}
}

// TemplateCommonData is a struct which is used for many templated files.
type TemplateCommonData struct {
	AssetPath              string
	Base                   string
	BaseURL                string
	CSPNonce               string
	LogoOverride           string
	DuoSelfEnrollment      string
	RememberMe             string
	ResetPassword          string
	ResetPasswordCustomURL string
	Session                string
	Theme                  string
}

func (data *TemplateCommonData) toMap() map[string]any {
	var res map[string]any

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	if err := json.Unmarshal(jsonData, &res); err != nil {
		return nil
	}

	return res
}
