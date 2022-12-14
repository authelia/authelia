package server

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ServeTemplatedFile serves a templated version of a specified file,
// this is utilised to pass information between the backend and frontend
// and generate a nonce to support a restrictive CSP while using material-ui.
func ServeTemplatedFile(publicDir, file string, opts *TemplatedFileOptions) middlewares.RequestHandler {
	logger := logging.Logger()

	a, err := assets.Open(path.Join(publicDir, file))
	if err != nil {
		logger.Fatalf("Unable to open %s: %s", file, err)
	}

	b, err := io.ReadAll(a)
	if err != nil {
		logger.Fatalf("Unable to read %s: %s", file, err)
	}

	tmpl, err := template.New("file").Parse(string(b))
	if err != nil {
		logger.Fatalf("Unable to parse %s template: %s", file, err)
	}

	isDevEnvironment := os.Getenv(environment) == dev

	return func(ctx *middlewares.AutheliaCtx) {
		logoOverride := f

		if opts.AssetPath != "" {
			if _, err = os.Stat(filepath.Join(opts.AssetPath, fileLogo)); err == nil {
				logoOverride = t
			}
		}

		switch extension := filepath.Ext(file); extension {
		case ".html", ".htm":
			ctx.SetContentTypeTextHTML()
		case ".json":
			ctx.SetContentTypeApplicationJSON()
		default:
			ctx.SetContentTypeTextPlain()
		}

		nonce := utils.RandomString(32, utils.CharSetAlphaNumeric, true)

		switch {
		case publicDir == assetsSwagger:
			ctx.Response.Header.Add(fasthttp.HeaderContentSecurityPolicy, fmt.Sprintf(tmplCSPSwagger, nonce, nonce))
		case ctx.Configuration.Server.Headers.CSPTemplate != "":
			ctx.Response.Header.Add(fasthttp.HeaderContentSecurityPolicy, strings.ReplaceAll(ctx.Configuration.Server.Headers.CSPTemplate, placeholderCSPNonce, nonce))
		case isDevEnvironment:
			ctx.Response.Header.Add(fasthttp.HeaderContentSecurityPolicy, fmt.Sprintf(tmplCSPDevelopment, nonce))
		default:
			ctx.Response.Header.Add(fasthttp.HeaderContentSecurityPolicy, fmt.Sprintf(tmplCSPDefault, nonce))
		}

		common := opts.CommonData(ctx.BasePath(), ctx.RootURLSlash().String(), nonce, logoOverride)

		ctx.Logger.WithField("base", common.Base).WithField("baseURL", common.BaseURL).WithField("proto", ctx.Request.Header.Protocol()).WithField("headers", ctx.Request.Header.String()).Debugf("serving templated file")

		if err = tmpl.Execute(ctx.Response.BodyWriter(), common); err != nil {
			ctx.RequestCtx.Error("an error occurred", 503)
			logger.Errorf("Unable to execute template: %v", err)

			return
		}
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

func NewTemplatedFileOptions(config *schema.Configuration) (opts *TemplatedFileOptions) {
	opts = &TemplatedFileOptions{
		AssetPath:              config.Server.AssetPath,
		DuoSelfEnrollment:      f,
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

// TemplatedFileOptions is a struct which is used for many templated files.
type TemplatedFileOptions struct {
	AssetPath              string
	DuoSelfEnrollment      string
	RememberMe             string
	ResetPassword          string
	ResetPasswordCustomURL string
	Session                string
	Theme                  string
}

// CommonData returns a TemplatedFileCommonData with the dynamic options.
func (options *TemplatedFileOptions) CommonData(base, baseURL, nonce, logoOverride string) TemplatedFileCommonData {
	return TemplatedFileCommonData{
		Base:                   base,
		BaseURL:                baseURL,
		CSPNonce:               nonce,
		LogoOverride:           logoOverride,
		DuoSelfEnrollment:      options.DuoSelfEnrollment,
		RememberMe:             options.RememberMe,
		ResetPassword:          options.ResetPassword,
		ResetPasswordCustomURL: options.ResetPasswordCustomURL,
		Session:                options.Session,
		Theme:                  options.Theme,
	}
}

// TemplatedFileCommonData is a struct which is used for many templated files.
type TemplatedFileCommonData struct {
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
