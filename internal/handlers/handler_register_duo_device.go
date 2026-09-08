package handlers

import (
	"strings"

	"github.com/authelia/authelia/v4/internal/duo"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// DuoDevicesGET handler for retrieving available devices and capabilities from Duo API.
func DuoDevicesGET(duoAPI duo.Provider) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		userSession, err := ctx.GetSession()
		if err != nil {
			ctx.GetLogger().WithError(err).Error(errStrUserSessionData)
			ctx.SetJSONError(messageMFAValidationFailed)

			return
		}

		ctx.Logger.Debugf("Starting Duo PreAuth for %s", userSession.Username)

		result, message, devices, enrollURL, err := DuoPreAuth(ctx, &userSession, duoAPI)
		if err != nil {
			ctx.GetLogger().WithError(err).Error("Error occurred performing the Duo PreAuth API call")
			ctx.SetJSONError(messageMFAValidationFailed)

			return
		}

		response := DuoDevicesResponse{}

		switch result {
		case auth:
			if devices == nil {
				ctx.Logger.Debugf("No applicable device/method available for Duo user %s", userSession.Username)

				response.Result = enroll
			} else {
				response.Result = auth
				response.Devices = devices
			}

			SendDuoDevicesResponse(ctx, response)

		case allow:
			ctx.Logger.Debugf("Device selection not possible for user %s, because Duo authentication was bypassed - Defaults to Auto Push", userSession.Username)

			response.Result = allow
			SendDuoDevicesResponse(ctx, response)

		case enroll:
			ctx.Logger.Debugf("Duo user: %s not enrolled", userSession.Username)

			response.Result = enroll
			response.EnrollURL = enrollURL
			SendDuoDevicesResponse(ctx, response)

		case deny:
			ctx.Logger.Debugf("Duo User not allowed to authenticate: %s", userSession.Username)

			response.Result = deny
			SendDuoDevicesResponse(ctx, response)

		default:
			ctx.GetLogger().Errorf("Error occurred performing the Duo PreAuth API call for user '%s' which returned the result '%s' with the message '%s'", userSession.Username, result, message)
			ctx.SetJSONError(messageMFAValidationFailed)
		}
	}
}

// DuoDevicePOST update the user preferences regarding Duo device and method.
func DuoDevicePOST(ctx *middlewares.AutheliaCtx) {
	bodyJSON := DuoDeviceBody{}

	var (
		userSession session.UserSession
		err         error
	)
	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred parsing the preferred Duo device request body")
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if !utils.IsStringInSlice(bodyJSON.Method, duo.PossibleMethods) {
		ctx.GetLogger().Errorf("Error occurred setting the preferred Duo device as the method '%s' is unknown, it should be one of %s", bodyJSON.Method, strings.Join(duo.PossibleMethods, ", "))
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.GetLogger().WithError(err).Error(errStrUserSessionData)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	ctx.Logger.Debugf("Save new preferred Duo device and method of user %s to %s using %s", userSession.Username, bodyJSON.Device, bodyJSON.Method)

	err = ctx.Providers.StorageProvider.SavePreferredDuoDevice(ctx, model.DuoDevice{Username: userSession.Username, Device: bodyJSON.Device, Method: bodyJSON.Method})
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred saving the new preferred Duo device and method")
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	ctx.ReplyOK()
}

// DuoDeviceDELETE deletes the users preferred Duo device and method.
func DuoDeviceDELETE(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)
	if userSession, err = ctx.GetSession(); err != nil {
		ctx.GetLogger().WithError(err).Error(errStrUserSessionData)
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	ctx.Logger.Debugf("Deleting preferred Duo device and method of user %s", userSession.Username)

	if err = ctx.Providers.StorageProvider.DeletePreferredDuoDevice(ctx, userSession.Username); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred deleting the preferred Duo device and method")
		ctx.SetJSONError(messageMFAValidationFailed)

		return
	}

	ctx.ReplyOK()
}

// SendDuoDevicesResponse sends a JSON response for Duo device operations.
func SendDuoDevicesResponse(ctx *middlewares.AutheliaCtx, response DuoDevicesResponse) {
	if err := ctx.SetJSONBody(response); err != nil {
		ctx.GetLogger().WithError(err).Error(errStrRespBody)
		ctx.SetJSONError(messageMFAValidationFailed)
	}
}
