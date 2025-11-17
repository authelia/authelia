package duo

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"

	duoapi "github.com/duosecurity/duo_api_golang"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

type Context interface {
	context.Context

	GetLogger() *logrus.Entry
	RemoteIP() net.IP
}

// BaseProvider describes the upstream provider we intend to utilize. Implemented by duoapi.
type BaseProvider interface {
	SignedCall(method string, uri string, params url.Values, options ...duoapi.DuoApiOption) (*http.Response, []byte, error)
	Call(method string, uri string, params url.Values, options ...duoapi.DuoApiOption) (*http.Response, []byte, error)
}

// The Provider interface is used to describe this provider for the purpose of mock testing.
type Provider interface {
	model.StartupCheck

	PreAuthCall(ctx Context, userSession *session.UserSession, values url.Values) (response *PreAuthResponse, err error)
	AuthCall(ctx Context, userSession *session.UserSession, values url.Values) (response *AuthResponse, err error)
}

// Production implementation of the Provider interface.
type Production struct {
	BaseProvider
}

// Device holds all necessary info for frontend.
type Device struct {
	Capabilities []string `json:"capabilities"`
	Device       string   `json:"device"`
	DisplayName  string   `json:"display_name"`
	Name         string   `json:"name"`
	SmsNextcode  string   `json:"sms_nextcode"`
	Number       string   `json:"number"`
	Type         string   `json:"type"`
}

// Response coming from Duo API.
type Response struct {
	Response      json.RawMessage `json:"response"`
	Code          int             `json:"code"`
	Message       string          `json:"message"`
	MessageDetail string          `json:"message_detail"`
	Stat          string          `json:"stat"`
}

// AuthResponse is a response for a authorization request.
type AuthResponse struct {
	Result             string `json:"result"`
	Status             string `json:"status"`
	StatusMessage      string `json:"status_msg"`
	TrustedDeviceToken string `json:"trusted_device_token"`
}

// PreAuthResponse is a response for a preauthorization request.
type PreAuthResponse struct {
	Result          string   `json:"result"`
	StatusMessage   string   `json:"status_msg"`
	Devices         []Device `json:"devices"`
	EnrollPortalURL string   `json:"enroll_portal_url"`
}
