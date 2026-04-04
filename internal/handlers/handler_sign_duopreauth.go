package handlers

import (
	"net/url"

	"github.com/authelia/authelia/v4/internal/duo"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// DuoPreAuth helper function for retrieving supported devices and capabilities from Duo API.
func DuoPreAuth(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, duoAPI duo.Provider) (result, message string, devices []DuoDevice, enrollURL string, err error) {
	values := url.Values{}
	values.Set("username", userSession.Username)

	preAuthResponse, err := duoAPI.PreAuthCall(ctx, userSession, values)
	if err != nil {
		return "", "", nil, "", err
	}

	if preAuthResponse.Result != auth {
		return preAuthResponse.Result, preAuthResponse.StatusMessage, nil, preAuthResponse.EnrollPortalURL, nil
	}

	supportedDevices := FilterSupportedDevices(preAuthResponse.Devices)
	if len(supportedDevices) == 0 {
		return preAuthResponse.Result, preAuthResponse.StatusMessage, nil, preAuthResponse.EnrollPortalURL, nil
	}

	return preAuthResponse.Result, preAuthResponse.StatusMessage, supportedDevices, preAuthResponse.EnrollPortalURL, nil
}

// FilterSupportedDevices filters Duo devices to only include those with supported methods.
func FilterSupportedDevices(duoDevices []duo.Device) []DuoDevice {
	var supportedDevices []DuoDevice

	for _, device := range duoDevices {
		supportedMethods := FilterSupportedMethods(device.Capabilities)
		if len(supportedMethods) > 0 {
			supportedDevices = append(supportedDevices, DuoDevice{
				Device:       device.Device,
				DisplayName:  device.DisplayName,
				Capabilities: supportedMethods,
			})
		}
	}

	return supportedDevices
}

// FilterSupportedMethods filters method capabilities to only include supported Duo methods.
func FilterSupportedMethods(capabilities []string) []string {
	var supportedMethods []string

	for _, method := range duo.PossibleMethods {
		if utils.IsStringInSlice(method, capabilities) {
			supportedMethods = append(supportedMethods, method)
		}
	}

	return supportedMethods
}
