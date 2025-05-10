package templates

import (
	"fmt"
	th "html/template"
	"io"
	tt "text/template"
	"time"

	"github.com/avct/uasurfer"

	"github.com/authelia/authelia/v4/internal/logging"

	"github.com/authelia/authelia/v4/internal/utils"
)

// Templates is the struct which holds all the *template.Template values.
type Templates struct {
	notification NotificationTemplates
	asset        AssetTemplates
	oidc         OpenIDConnectTemplates
}

type OpenIDConnectTemplates struct {
	formpost *th.Template
}

// AssetTemplates are templates for specific key assets.
type AssetTemplates struct {
	index *tt.Template
	api   OpenAPIAssetTemplates
}

// OpenAPIAssetTemplates are asset templates for the OpenAPI specification.
type OpenAPIAssetTemplates struct {
	index *tt.Template
	spec  *tt.Template
}

// NotificationTemplates are the templates for the notification system.
type NotificationTemplates struct {
	jwtIdentityVerification *EmailTemplate
	otcIdentityVerification *EmailTemplate
	event                   *EmailTemplate
	newLogin                *EmailTemplate
}

// Template covers shared implementations between the text and html template.Template.
type Template interface {
	Execute(wr io.Writer, data any) error
	ExecuteTemplate(wr io.Writer, name string, data any) error
	Name() string
	DefinedTemplates() string
}

// Config for the Provider.
type Config struct {
	EmailTemplatesPath string
}

// EmailTemplate is the template type which contains both the html and txt versions of a template.
type EmailTemplate struct {
	HTML *th.Template
	Text *tt.Template
}

// EmailEventValues are the values used for event templates.
type EmailEventValues struct {
	Title       string
	BodyPrefix  string
	BodySuffix  string
	BodyEvent   string
	DisplayName string
	Details     map[string]any
	RemoteIP    string
}

// EmailNewLoginValues are the values used for new login templates.
type EmailNewLoginValues struct {
	Title        string
	Date         string
	Domain       string
	DisplayName  string
	RemoteIP     string
	DeviceInfo   string
	RawUserAgent string
}

func NewEmailNewLoginValues(displayName, domain, remoteIP string, userAgent *uasurfer.UserAgent, rawUserAgent string, timestamp time.Time) EmailNewLoginValues {
	unknown := "Unknown"
	formattedDate := timestamp.Format("Monday, January 2, 2006 at 3:04:05 PM -07:00")

	browserName := userAgent.Browser.Name.StringTrimPrefix()
	browserVersion := utils.FormatVersion(userAgent.Browser.Version)
	osName := userAgent.OS.Name.StringTrimPrefix()
	osVersion := utils.FormatVersion(userAgent.OS.Version)
	deviceType := userAgent.DeviceType.StringTrimPrefix()

	browserStr := browserName
	if browserVersion != "" && browserVersion != unknown {
		browserStr = fmt.Sprintf("%s %s", browserName, browserVersion)
	}

	osStr := osName
	if osVersion != "" && osVersion != unknown {
		osStr = fmt.Sprintf("%s %s", osName, osVersion)
	}

	deviceInfo := fmt.Sprintf("%s on %s", browserStr, osStr)
	if deviceType != "" {
		deviceInfo = fmt.Sprintf("%s (%s)", deviceInfo, deviceType)
	}

	logging.Logger().Debug(deviceInfo)

	return EmailNewLoginValues{
		Title:        "Login From New IP",
		Date:         formattedDate,
		Domain:       domain,
		DisplayName:  displayName,
		RemoteIP:     remoteIP,
		DeviceInfo:   deviceInfo,
		RawUserAgent: rawUserAgent,
	}
}

// EmailIdentityVerificationJWTValues are the values used for the identity verification JWT templates.
type EmailIdentityVerificationJWTValues struct {
	Title              string
	DisplayName        string
	Domain             string
	RemoteIP           string
	LinkURL            string
	LinkText           string
	RevocationLinkURL  string
	RevocationLinkText string
}

// EmailIdentityVerificationOTCValues are the values used for the identity verification OTP templates.
type EmailIdentityVerificationOTCValues struct {
	Title              string
	DisplayName        string
	Domain             string
	RemoteIP           string
	OneTimeCode        string
	RevocationLinkURL  string
	RevocationLinkText string
}
