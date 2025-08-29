package handlers

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/v4/internal/duo"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// DuoPOST handler for sending a push notification via duo api.
func DuoPOST(duoAPI duo.Provider) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		var (
			bodyJSON       = &bodySignDuoRequest{}
			device, method string

			userSession session.UserSession
			err         error
		)
		if err = ctx.ParseBody(bodyJSON); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrParseRequestBody, regulation.AuthTypeDuo)

			respondUnauthorized(ctx, messageMFAValidationFailed)

			return
		}

		if userSession, err = ctx.GetSession(); err != nil {
			ctx.Error(fmt.Errorf("error occurred retrieving user session: %w", err), messageMFAValidationFailed)
			return
		}

		remoteIP := ctx.RemoteIP().String()

		duoDevice, err := ctx.Providers.StorageProvider.LoadPreferredDuoDevice(ctx, userSession.Username)
		if err != nil {
			ctx.Logger.Debugf("Error identifying preferred device for user %s: %s", userSession.Username, err)
			ctx.Logger.Debugf("Starting Duo PreAuth for initial device selection of user: %s", userSession.Username)
			device, method, err = HandleInitialDeviceSelection(ctx, &userSession, duoAPI, bodyJSON)
		} else {
			ctx.Logger.Debugf("Starting Duo PreAuth to check preferred device of user: %s", userSession.Username)
			device, method, err = HandlePreferredDeviceCheck(ctx, &userSession, duoAPI, duoDevice.Device, duoDevice.Method, bodyJSON)
		}

		if err != nil {
			ctx.Error(err, messageMFAValidationFailed)
			return
		}

		if device == "" || method == "" {
			return
		}

		ctx.Logger.Debugf("Starting Duo Auth attempt for %s with device %s and method %s from IP %s", userSession.Username, device, method, remoteIP)

		values, err := SetValues(userSession, device, method, remoteIP, bodyJSON.TargetURL, bodyJSON.Passcode)
		if err != nil {
			ctx.Logger.Errorf("Failed to set values for Duo Auth Call for user '%s': %+v", userSession.Username, err)

			respondUnauthorized(ctx, messageMFAValidationFailed)

			return
		}

		authResponse, err := duoAPI.AuthCall(ctx, &userSession, values)
		if err != nil {
			ctx.Logger.Errorf("Failed to perform Duo Auth Call for user '%s': %+v", userSession.Username, err)

			respondUnauthorized(ctx, messageMFAValidationFailed)

			return
		}

		if authResponse.Result != allow {
			doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, userSession.Username, nil), regulation.AuthTypeDuo,
				fmt.Errorf("duo auth result: %s, status: %s, message: %s", authResponse.Result, authResponse.Status,
					authResponse.StatusMessage))

			respondUnauthorized(ctx, messageMFAValidationFailed)

			return
		}

		doMarkAuthenticationAttempt(ctx, true, regulation.NewBan(regulation.BanTypeNone, userSession.Username, nil), regulation.AuthTypeDuo, nil)

		HandleAllow(ctx, &userSession, bodyJSON)
	}
}

// HandleInitialDeviceSelection handler for retrieving all available devices.
func HandleInitialDeviceSelection(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, duoAPI duo.Provider, bodyJSON *bodySignDuoRequest) (device string, method string, err error) {
	result, message, devices, enrollURL, err := DuoPreAuth(ctx, userSession, duoAPI)
	if err != nil {
		ctx.Logger.Errorf("Failed to perform Duo PreAuth for user '%s': %+v", userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return "", "", err
	}

	switch result {
	case enroll:
		ctx.Logger.Debugf("Duo user: %s not enrolled", userSession.Username)

		if err := ctx.SetJSONBody(DuoSignResponse{Result: enroll, EnrollURL: enrollURL}); err != nil {
			return "", "", fmt.Errorf("unable to set JSON body in response")
		}

		return "", "", nil
	case deny:
		ctx.Logger.Infof("Duo user: %s not allowed to authenticate: %s", userSession.Username, message)

		if err := ctx.SetJSONBody(DuoSignResponse{Result: deny}); err != nil {
			return "", "", fmt.Errorf("unable to set JSON body in response")
		}

		return "", "", nil
	case allow:
		ctx.Logger.Debugf("Duo authentication was bypassed for user: %s", userSession.Username)
		HandleAllow(ctx, userSession, bodyJSON)

		return "", "", nil
	case auth:
		device, method, err = HandleAutoSelection(ctx, devices, userSession.Username)
		if err != nil {
			return "", "", err
		}

		return device, method, nil
	}

	return "", "", fmt.Errorf("unknown result: %s", result)
}

// HandlePreferredDeviceCheck handler to check if the saved device and method is still valid.
func HandlePreferredDeviceCheck(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, duoAPI duo.Provider, device string, method string, bodyJSON *bodySignDuoRequest) (string, string, error) {
	result, message, devices, enrollURL, err := DuoPreAuth(ctx, userSession, duoAPI)
	if err != nil {
		ctx.Logger.Errorf("Failed to perform Duo PreAuth for user '%s': %+v", userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return "", "", nil
	}

	switch result {
	case enroll:
		ctx.Logger.Debugf("Duo user: %s no longer enrolled removing preferred device", userSession.Username)

		if err := ctx.Providers.StorageProvider.DeletePreferredDuoDevice(ctx, userSession.Username); err != nil {
			return "", "", fmt.Errorf("unable to delete preferred Duo device and method for user %s: %w", userSession.Username, err)
		}

		if err := ctx.SetJSONBody(DuoSignResponse{Result: enroll, EnrollURL: enrollURL}); err != nil {
			return "", "", fmt.Errorf("unable to set JSON body in response")
		}

		return "", "", nil
	case deny:
		ctx.Logger.Infof("Duo user: %s not allowed to authenticate: %s", userSession.Username, message)
		ctx.ReplyUnauthorized()

		return "", "", nil
	case allow:
		ctx.Logger.Debugf("Duo authentication was bypassed for user: %s", userSession.Username)
		HandleAllow(ctx, userSession, bodyJSON)

		return "", "", nil
	case auth:
		if devices == nil {
			ctx.Logger.Debugf("Duo user: %s has no compatible device/method available removing preferred device", userSession.Username)

			if err := ctx.Providers.StorageProvider.DeletePreferredDuoDevice(ctx, userSession.Username); err != nil {
				return "", "", fmt.Errorf("unable to delete preferred Duo device and method for user %s: %w", userSession.Username, err)
			}

			if err := ctx.SetJSONBody(DuoSignResponse{Result: enroll}); err != nil {
				return "", "", fmt.Errorf("unable to set JSON body in response")
			}

			return "", "", nil
		}

		if len(devices) > 0 {
			for i := range devices {
				if devices[i].Device == device {
					if utils.IsStringInSlice(method, devices[i].Capabilities) {
						return device, method, nil
					}
				}
			}
		}

		return HandleAutoSelection(ctx, devices, userSession.Username)
	}

	return "", "", fmt.Errorf("unknown result: %s", result)
}

// HandleAutoSelection handler automatically selects preferred device if there is only one suitable option.
func HandleAutoSelection(ctx *middlewares.AutheliaCtx, devices []DuoDevice, username string) (string, string, error) {
	if devices == nil {
		ctx.Logger.Debugf("No compatible device/method available for Duo user: %s", username)

		if err := ctx.SetJSONBody(DuoSignResponse{Result: enroll}); err != nil {
			return "", "", fmt.Errorf("unable to set JSON body in response")
		}

		return "", "", nil
	}

	if len(devices) > 1 {
		ctx.Logger.Debugf("Multiple devices available for Duo user: %s require manual selection", username)

		if err := ctx.SetJSONBody(DuoSignResponse{Result: auth, Devices: devices}); err != nil {
			return "", "", fmt.Errorf("unable to set JSON body in response")
		}

		return "", "", nil
	}

	if len(devices[0].Capabilities) > 1 {
		ctx.Logger.Debugf("Multiple methods available for Duo user: %s require manual selection", username)

		if err := ctx.SetJSONBody(DuoSignResponse{Result: auth, Devices: devices}); err != nil {
			return "", "", fmt.Errorf("unable to set JSON body in response")
		}

		return "", "", nil
	}

	device := devices[0].Device
	method := devices[0].Capabilities[0]
	ctx.Logger.Debugf("Exactly one device: '%s' and method: '%s' found, saving as new preferred Duo device and method for user: %s", device, method, username)

	if err := ctx.Providers.StorageProvider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: username, Method: method, Device: device}); err != nil {
		return "", "", fmt.Errorf("unable to save new preferred Duo device and method for user %s: %w", username, err)
	}

	return device, method, nil
}

// HandleAllow handler for successful logins.
func HandleAllow(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, bodyJSON *bodySignDuoRequest) {
	var (
		err error
	)
	if err = ctx.RegenerateSession(); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrSessionRegenerate, regulation.AuthTypeDuo, userSession.Username)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	userSession.SetTwoFactorDuo(ctx.GetClock().Now())

	if err = ctx.SaveSession(*userSession); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "authentication time", regulation.AuthTypeTOTP, logFmtActionAuthentication, userSession.Username)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if len(bodyJSON.Flow) > 0 {
		handleFlowResponse(ctx, userSession, bodyJSON.FlowID, bodyJSON.Flow, bodyJSON.SubFlow, bodyJSON.UserCode)
	} else {
		Handle2FAResponse(ctx, bodyJSON.TargetURL)
	}
}

// SetValues sets all appropriate Values for the Auth Request.
func SetValues(userSession session.UserSession, device string, method string, remoteIP string, targetURL string, passcode string) (url.Values, error) {
	values := url.Values{}
	values.Set("username", userSession.Username)
	values.Set("ipaddr", remoteIP)
	values.Set("factor", method)

	switch method {
	case duo.Push:
		values.Set("device", device)

		if userSession.DisplayName != "" {
			values.Set("display_username", userSession.DisplayName)
		}

		if targetURL != "" {
			values.Set("pushinfo", fmt.Sprintf("target%%20url=%s", targetURL))
		}
	case duo.Phone:
		values.Set("device", device)
	case duo.SMS:
		values.Set("device", device)
	case duo.OTP:
		if passcode != "" {
			values.Set("passcode", passcode)
		} else {
			return nil, fmt.Errorf("no passcode received from user: %s", userSession.Username)
		}
	}

	return values, nil
}
