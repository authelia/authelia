package handlers

import (
	"net/url"

	"github.com/authelia/authelia/v4/internal/duo"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// DuoPreAuth helper function for retrieving supported devices and capabilities from duo api.
func DuoPreAuth(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, duoAPI duo.Provider) (result, message string, devices []DuoDevice, enrollURL string, err error) {
	values := url.Values{}
	values.Set("username", userSession.Username)

	preAuthResponse, err := duoAPI.PreAuthCall(ctx, userSession, values)
	if err != nil {
		return "", "", nil, "", err
	}

	if preAuthResponse.Result == auth {
		var supportedDevices []DuoDevice

		for _, device := range preAuthResponse.Devices {
			var supportedMethods []string

			for _, method := range duo.PossibleMethods {
				if utils.IsStringInSlice(method, device.Capabilities) {
					supportedMethods = append(supportedMethods, method)
				}
			}

			if len(supportedMethods) > 0 {
				supportedDevices = append(supportedDevices, DuoDevice{
					Device:       device.Device,
					DisplayName:  device.DisplayName,
					Capabilities: supportedMethods,
				})
			}
		}

		if len(supportedDevices) > 0 {
			return preAuthResponse.Result, preAuthResponse.StatusMessage, supportedDevices, preAuthResponse.EnrollPortalURL, nil
		}
	}

	return preAuthResponse.Result, preAuthResponse.StatusMessage, nil, preAuthResponse.EnrollPortalURL, nil
}
