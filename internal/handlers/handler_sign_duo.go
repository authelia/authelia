package handlers

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/authelia/authelia/v4/internal/duo"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// DuoGET handler for retrieving the preferred Duo device.
func DuoGET(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Error(fmt.Errorf("%s: %w", errStrUserSessionData, err), messageMFAValidationFailed)
		return
	}

	duoDevice, err := ctx.Providers.StorageProvider.LoadPreferredDuoDevice(ctx, userSession.Username)
	if err != nil {
		ctx.Logger.Debugf("No preferred Duo device found for user: '%s': %v", userSession.Username, err)

		if err := ctx.SetJSONBody(DuoDevicesResponse{Result: auth}); err != nil {
			ctx.Error(errors.New(errStrRespBody), messageMFAValidationFailed)
		}

		return
	}

	ctx.Logger.Debugf("Found preferred Duo device '%s' and method '%s' for user: '%s'", duoDevice.Device, duoDevice.Method, userSession.Username)

	if err := ctx.SetJSONBody(DuoDevicesResponse{Result: auth, PreferredDevice: duoDevice.Device, PreferredMethod: duoDevice.Method}); err != nil {
		ctx.Error(errors.New(errStrRespBody), messageMFAValidationFailed)
	}
}

// DuoPOST handler for sending a push notification via Duo API.
func DuoPOST(duoAPI duo.Provider) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		bodyJSON := &bodySignDuoRequest{}
		if err := ctx.ParseBody(bodyJSON); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrParseRequestBody, regulation.AuthTypeDuo)
			respondUnauthorized(ctx, messageMFAValidationFailed)

			return
		}

		userSession, err := ctx.GetSession()
		if err != nil {
			ctx.Error(fmt.Errorf("%s: %w", errStrUserSessionData, err), messageMFAValidationFailed)
			return
		}

		if err := HandleDuoAuthentication(ctx, &userSession, duoAPI, bodyJSON); err != nil {
			respondUnauthorized(ctx, messageMFAValidationFailed)
			return
		}
	}
}

// PerformDuoAuthentication executes the Duo authentication call.
func PerformDuoAuthentication(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, duoAPI duo.Provider, device, method, remoteIP string, bodyJSON *bodySignDuoRequest) error {
	ctx.Logger.Debugf("Starting Duo Auth attempt for '%s' with device '%s' and method '%s' from IP '%s'", userSession.Username, device, method, remoteIP)

	values, err := SetValues(*userSession, device, method, remoteIP, bodyJSON.TargetURL, bodyJSON.Passcode)
	if err != nil {
		ctx.Logger.Errorf("Failed to set values for Duo Auth Call for user '%s': %+v", userSession.Username, err)
		return err
	}

	authResponse, err := duoAPI.AuthCall(ctx, userSession, values)
	if err != nil {
		ctx.Logger.Errorf("Failed to perform Duo Auth Call for user '%s': %+v", userSession.Username, err)
		return err
	}

	if authResponse.Result != allow {
		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, userSession.Username, nil), regulation.AuthTypeDuo,
			fmt.Errorf("duo auth result: '%s', status: '%s', message: '%s'", authResponse.Result, authResponse.Status,
				authResponse.StatusMessage))

		return fmt.Errorf("duo authentication failed")
	}

	doMarkAuthenticationAttempt(ctx, true, regulation.NewBan(regulation.BanTypeNone, userSession.Username, nil), regulation.AuthTypeDuo, nil)

	return nil
}

// PerformDuoAuthenticationIfValid performs Duo authentication if device and method are valid.
func PerformDuoAuthenticationIfValid(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, duoAPI duo.Provider, device, method string, bodyJSON *bodySignDuoRequest) error {
	if device == "" || method == "" {
		return nil
	}

	remoteIP := ctx.RemoteIP().String()
	if err := PerformDuoAuthentication(ctx, userSession, duoAPI, device, method, remoteIP, bodyJSON); err != nil {
		return err
	}

	HandleAllow(ctx, userSession, bodyJSON)

	return nil
}

// HandleInitialDeviceSelection handler for retrieving all available devices.
func HandleInitialDeviceSelection(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, duoAPI duo.Provider, bodyJSON *bodySignDuoRequest) (device string, method string, err error) {
	result, message, devices, enrollURL, err := DuoPreAuth(ctx, userSession, duoAPI)
	if err != nil {
		ctx.Logger.Errorf("Failed to perform Duo PreAuth for user '%s': %+v", userSession.Username, err)
		respondUnauthorized(ctx, messageMFAValidationFailed)

		return "", "", err
	}

	return HandleDuoPreAuthResult(ctx, userSession, result, message, devices, enrollURL, bodyJSON)
}

// HandleDuoAuthentication handles the complete Duo authentication flow.
func HandleDuoAuthentication(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, duoAPI duo.Provider, bodyJSON *bodySignDuoRequest) error {
	duoDevice, err := ctx.Providers.StorageProvider.LoadPreferredDuoDevice(ctx, userSession.Username)
	if err != nil {
		selectedDevice, selectedMethod, err := HandleInitialDeviceSelection(ctx, userSession, duoAPI, bodyJSON)
		if err != nil {
			return err
		}

		return PerformDuoAuthenticationIfValid(ctx, userSession, duoAPI, selectedDevice, selectedMethod, bodyJSON)
	}

	selectedDevice, selectedMethod, err := HandlePreferredDeviceCheck(ctx, userSession, duoAPI, duoDevice.Device, duoDevice.Method, bodyJSON)
	if err != nil {
		return err
	}

	return PerformDuoAuthenticationIfValid(ctx, userSession, duoAPI, selectedDevice, selectedMethod, bodyJSON)
}

// HandleDuoPreAuthResult processes the result of a DuoPreAuth call and handles common response logic.
func HandleDuoPreAuthResult(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, result, message string, devices []DuoDevice, enrollURL string, bodyJSON *bodySignDuoRequest) (device, method string, err error) {
	switch result {
	case enroll:
		ctx.Logger.Debugf("Duo user: '%s' not enrolled", userSession.Username)

		if err := ctx.SetJSONBody(DuoSignResponse{Result: enroll, EnrollURL: enrollURL}); err != nil {
			return "", "", errors.New(errStrRespBody)
		}

		return "", "", nil

	case deny:
		ctx.Logger.Infof("Duo user: '%s' not allowed to authenticate: %s", userSession.Username, message)

		if err := ctx.SetJSONBody(DuoSignResponse{Result: deny}); err != nil {
			return "", "", errors.New(errStrRespBody)
		}

		return "", "", nil

	case allow:
		ctx.Logger.Debugf("Duo authentication was bypassed for user: '%s'", userSession.Username)
		HandleAllow(ctx, userSession, bodyJSON)

		return "", "", nil

	case auth:
		return HandleAuthResult(ctx, userSession, devices, "", "")

	default:
		return "", "", fmt.Errorf("unknown result: %s", result)
	}
}

// HandleNoDevicesAvailable handles the case when no compatible devices are available.
func HandleNoDevicesAvailable(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, storedDevice string) error {
	ctx.Logger.Debugf("No compatible device/method available for Duo user: '%s'", userSession.Username)

	if storedDevice != "" {
		if deleteErr := ctx.Providers.StorageProvider.DeletePreferredDuoDevice(ctx, userSession.Username); deleteErr != nil {
			return fmt.Errorf("unable to delete preferred Duo device and method for user '%s': %w", userSession.Username, deleteErr)
		}
	}

	if err := ctx.SetJSONBody(DuoSignResponse{Result: enroll}); err != nil {
		return errors.New(errStrRespBody)
	}

	return nil
}

// SelectDeviceFromAvailable selects the best device from available options.
func SelectDeviceFromAvailable(devices []DuoDevice, storedDevice, storedMethod string) (device, method string) {
	if storedDevice != "" && storedMethod != "" {
		if selectedDevice, selectedMethod := FindValidDevice(devices, storedDevice, storedMethod); selectedDevice != "" {
			return selectedDevice, selectedMethod
		}
	}

	return "", ""
}

// HandleAuthResult handles the auth result case for device selection.
func HandleAuthResult(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, devices []DuoDevice, storedDevice, storedMethod string) (device, method string, err error) {
	if devices == nil {
		if err := HandleNoDevicesAvailable(ctx, userSession, storedDevice); err != nil {
			return "", "", err
		}

		return "", "", nil
	}

	selectedDevice, selectedMethod := SelectDeviceFromAvailable(devices, storedDevice, storedMethod)
	if selectedDevice != "" {
		return selectedDevice, selectedMethod, nil
	}

	return HandleAutoSelection(ctx, devices, userSession.Username)
}

// FindValidDevice checks if a preferred device/method combination is still available.
func FindValidDevice(devices []DuoDevice, preferredDevice, preferredMethod string) (device, method string) {
	for _, dev := range devices {
		if dev.Device == preferredDevice && utils.IsStringInSlice(preferredMethod, dev.Capabilities) {
			return preferredDevice, preferredMethod
		}
	}

	return "", ""
}

// HandlePreferredDeviceCheck handler to check if the saved device and method is still valid.
func HandlePreferredDeviceCheck(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, duoAPI duo.Provider, storedDevice string, storedMethod string, bodyJSON *bodySignDuoRequest) (string, string, error) {
	result, message, devices, enrollURL, err := DuoPreAuth(ctx, userSession, duoAPI)
	if err != nil {
		ctx.Logger.Errorf("Failed to perform Duo PreAuth for user '%s': %+v", userSession.Username, err)
		respondUnauthorized(ctx, messageMFAValidationFailed)

		return "", "", nil
	}

	switch result {
	case enroll:
		ctx.Logger.Debugf("Duo user: '%s' no longer enrolled removing preferred device", userSession.Username)

		if err := ctx.Providers.StorageProvider.DeletePreferredDuoDevice(ctx, userSession.Username); err != nil {
			return "", "", fmt.Errorf("unable to delete preferred Duo device and method for user '%s': %w", userSession.Username, err)
		}

		if err := ctx.SetJSONBody(DuoSignResponse{Result: enroll, EnrollURL: enrollURL}); err != nil {
			return "", "", errors.New(errStrRespBody)
		}

		return "", "", nil
	case deny:
		ctx.Logger.Infof("Duo user: '%s' not allowed to authenticate: %s", userSession.Username, message)
		ctx.ReplyUnauthorized()

		return "", "", nil
	case allow:
		ctx.Logger.Debugf("Duo authentication was bypassed for user: '%s'", userSession.Username)
		HandleAllow(ctx, userSession, bodyJSON)

		return "", "", nil
	case auth:
		return HandleAuthResult(ctx, userSession, devices, storedDevice, storedMethod)

	default:
		return "", "", fmt.Errorf("unknown result: %s", result)
	}
}

// HandleAutoSelection handler automatically selects preferred device if there is only one suitable option.
func HandleAutoSelection(ctx *middlewares.AutheliaCtx, devices []DuoDevice, username string) (string, string, error) {
	if devices == nil {
		ctx.Logger.Debugf("No compatible device/method available for Duo user: '%s'", username)

		if err := ctx.SetJSONBody(DuoSignResponse{Result: enroll}); err != nil {
			return "", "", errors.New(errStrRespBody)
		}

		return "", "", nil
	}

	if len(devices) > 1 {
		ctx.Logger.Debugf("Multiple devices available for Duo user: '%s' require manual selection", username)

		if err := ctx.SetJSONBody(DuoSignResponse{Result: auth, Devices: devices}); err != nil {
			return "", "", errors.New(errStrRespBody)
		}

		return "", "", nil
	}

	if len(devices[0].Capabilities) > 1 {
		ctx.Logger.Debugf("Multiple methods available for Duo user: '%s' require manual selection", username)

		if err := ctx.SetJSONBody(DuoSignResponse{Result: auth, Devices: devices}); err != nil {
			return "", "", errors.New(errStrRespBody)
		}

		return "", "", nil
	}

	device := devices[0].Device
	method := devices[0].Capabilities[0]
	ctx.Logger.Debugf("Exactly one device: '%s' and method: '%s' found, saving as new preferred Duo device and method for user: '%s'", device, method, username)

	if err := ctx.Providers.StorageProvider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: username, Method: method, Device: device}); err != nil {
		return "", "", fmt.Errorf("unable to save new preferred Duo device and method for user '%s': %w", username, err)
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

	userSession.SetTwoFactorDuo(ctx.Clock.Now())

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
