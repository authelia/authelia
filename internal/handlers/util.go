package handlers

import (
	"bytes"
	"fmt"
	"net/url"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/templates"
)

var bytesEmpty = []byte("")

func ctxGetPortalURL(ctx *middlewares.AutheliaCtx) (portalURL *url.URL) {
	var rawURL []byte

	if rawURL = ctx.QueryArgRedirect(); rawURL != nil && !bytes.Equal(rawURL, bytesEmpty) {
		portalURL, _ = url.ParseRequestURI(string(rawURL))

		return portalURL
	} else if rawURL = ctx.XAutheliaURL(); rawURL != nil && !bytes.Equal(rawURL, bytesEmpty) {
		portalURL, _ = url.ParseRequestURI(string(rawURL))

		return portalURL
	}

	return nil
}

func ctxLogEvent(ctx *middlewares.AutheliaCtx, username, description string, eventDetails map[string]any) {
	var (
		details *authentication.UserDetails
		err     error
	)

	ctx.Logger.Debugf("Getting user details for notification")

	// Send Notification.
	if details, err = ctx.Providers.UserProvider.GetDetails(username); err != nil {
		ctx.Logger.Error(err)
		return
	}

	if len(details.Emails) == 0 {
		ctx.Logger.Error(fmt.Errorf("user %s has no email address configured", username))
		return
	}

	data := templates.EmailEventValues{
		Title:       description,
		DisplayName: details.DisplayName,
		RemoteIP:    ctx.RemoteIP().String(),
		Details:     eventDetails,
	}

	ctx.Logger.Debugf("Getting user addresses for notification")

	addresses := details.Addresses()

	ctx.Logger.Debugf("Sending an email to user %s (%s) to inform them of an important event.", username, addresses[0])

	if err = ctx.Providers.Notifier.Send(ctx, addresses[0], description, ctx.Providers.Templates.GetEventEmailTemplate(), data); err != nil {
		ctx.Logger.Error(err)
		return
	}
}
