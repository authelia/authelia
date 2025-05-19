package duo

import (
	"context"
	"encoding/json"
	"net"
	"net/url"

	duoapi "github.com/duosecurity/duo_api_golang"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/session"
)

type Context interface {
	context.Context

	GetLogger() *logrus.Entry
	RemoteIP() net.IP
}

// Provider interface wrapping duoapi library for testing purpose.
type Provider interface {
	Call(ctx Context, userSession *session.UserSession, values url.Values, method string, path string) (response *Response, err error)
	PreAuthCall(ctx Context, userSession *session.UserSession, values url.Values) (response *PreAuthResponse, err error)
	AuthCall(ctx Context, userSession *session.UserSession, values url.Values) (response *AuthResponse, err error)
}

// ProductionProvider implementation of Provider interface.
type ProductionProvider struct {
	*duoapi.DuoApi
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
