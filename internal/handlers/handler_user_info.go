package handlers

import (
	"fmt"
	"strings"
	"sync"

	"github.com/clems4ever/authelia/internal/authentication"
	"github.com/clems4ever/authelia/internal/middlewares"
	"github.com/clems4ever/authelia/internal/storage"
	"github.com/clems4ever/authelia/internal/utils"
	"github.com/sirupsen/logrus"
)

func loadInfo(username string, storageProvier storage.Provider, preferences *UserPreferences, logger *logrus.Entry) []error {
	var wg sync.WaitGroup
	wg.Add(3)

	errors := make([]error, 0)
	go func() {
		defer wg.Done()
		method, err := storageProvier.LoadPrefered2FAMethod(username)
		if err != nil {
			errors = append(errors, err)
			logger.Error(err)
			return
		}
		if method == "" {
			preferences.Method = authentication.PossibleMethods[0]
		} else {
			preferences.Method = method
		}
	}()

	go func() {
		defer wg.Done()
		_, _, err := storageProvier.LoadU2FDeviceHandle(username)
		if err != nil {
			if err == storage.ErrNoU2FDeviceHandle {
				return
			}
			errors = append(errors, err)
			logger.Error(err)
			return
		}
		preferences.HasU2F = true
	}()

	go func() {
		defer wg.Done()
		_, err := storageProvier.LoadTOTPSecret(username)
		if err != nil {
			if err == storage.ErrNoTOTPSecret {
				return
			}
			errors = append(errors, err)
			logger.Error(err)
			return
		}
		preferences.HasTOTP = true
	}()

	wg.Wait()
	return errors
}

// UserInfoGet get the info related to the user identitified by the session.
func UserInfoGet(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	preferences := UserPreferences{}
	errors := loadInfo(userSession.Username, ctx.Providers.StorageProvider, &preferences, ctx.Logger)

	if len(errors) > 0 {
		ctx.Error(fmt.Errorf("Unable to load user preferences"), operationFailedMessage)
		return
	}
	ctx.SetJSONBody(preferences)
}

type MethodBody struct {
	Method string `json:"method" valid:"required"`
}

// MethodPreferencePost update the user preferences regarding 2FA method.
func MethodPreferencePost(ctx *middlewares.AutheliaCtx) {
	bodyJSON := MethodBody{}
	err := ctx.ParseBody(&bodyJSON)
	if err != nil {
		ctx.Error(err, operationFailedMessage)
		return
	}

	if !utils.IsStringInSlice(bodyJSON.Method, authentication.PossibleMethods) {
		ctx.Error(fmt.Errorf("Unknown method '%s', it should be one of %s", bodyJSON.Method, strings.Join(authentication.PossibleMethods, ", ")), operationFailedMessage)
		return
	}

	userSession := ctx.GetSession()
	ctx.Logger.Debugf("Save new prefered 2FA method of user %s to %s", userSession.Username, bodyJSON.Method)
	err = ctx.Providers.StorageProvider.SavePrefered2FAMethod(userSession.Username, bodyJSON.Method)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to save new prefered 2FA method: %s", err), operationFailedMessage)
		return
	}

	ctx.ReplyOK()
}
