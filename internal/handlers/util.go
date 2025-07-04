package handlers

import (
	"fmt"
	"net"
	"net/mail"
	"strings"
	"time"

	"github.com/avct/uasurfer"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
)

const (
	eventLogKeyAction      = "Action"
	eventLogKeyCategory    = "Category"
	eventLogKeyDescription = "Description"

	eventEmailAction2FABody  = "Second Factor Method"
	eventLogAction2FAAdded   = "Second Factor Method Added"
	eventLogAction2FARemoved = "Second Factor Method Removed"

	eventEmailAction2FAPrefix        = "a"
	eventEmailAction2FAAddedSuffix   = "was added to your account."
	eventEmailAction2FARemovedSuffix = "was removed from your account."

	eventEmailActionPasswordModifyPrefix = "your"
	eventEmailActionPasswordReset        = "Password Reset"
	eventEmailActionPasswordChange       = "Password Change"
	eventEmailActionPasswordModifySuffix = "was successful."

	eventLogCategoryOneTimePassword    = "One-Time Password"
	eventLogCategoryWebAuthnCredential = "WebAuthn Credential" //nolint:gosec
)

type emailEventBody struct {
	Prefix string
	Body   string
	Suffix string
}

func ctxLogEvent(ctx *middlewares.AutheliaCtx, username, description string, body emailEventBody, eventDetails map[string]any) {
	var (
		details *authentication.UserDetails
		err     error
	)

	ctx.Logger.Debugf("Getting user details for notification")

	// Send Notification.
	if details, err = ctx.Providers.UserProvider.GetDetails(username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred looking up user details for user '%s' while attempting to alert them of an important event", username)
		return
	}

	if len(details.Emails) == 0 {
		ctx.Logger.WithError(fmt.Errorf("no email address was found for user")).Errorf("Error occurred looking up user details for user '%s' while attempting to alert them of an important event", username)
		return
	}

	data := templates.EmailEventValues{
		Title:       description,
		DisplayName: details.DisplayName,
		RemoteIP:    ctx.RemoteIP().String(),
		Details:     eventDetails,
		BodyPrefix:  body.Prefix,
		BodyEvent:   body.Body,
		BodySuffix:  body.Suffix,
	}

	ctx.Logger.Debugf("Getting user addresses for notification")

	addresses := details.Addresses()

	ctx.Logger.Debugf("Sending an email to user %s (%s) to inform them of an important event.", username, addresses[0].String())

	if err = ctx.Providers.Notifier.Send(ctx, addresses[0], description, ctx.Providers.Templates.GetEventEmailTemplate(), data); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred sending notification to user '%s' while attempting to alert them of an important event", username)
		return
	}
}

// redactEmail masks the local part of an email address for privacy, showing only first and last characters.
func redactEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}

	localPart := parts[0]
	domain := parts[1]

	if len(localPart) <= 2 {
		return strings.Repeat("*", len(localPart)) + "@" + domain
	}

	first := string(localPart[0])
	last := string(localPart[len(localPart)-1])
	middle := strings.Repeat("*", len(localPart)-2)

	return first + middle + last + "@" + domain
}

// IsIPTrusted checks if an IP address should be trusted based on private ranges and configured trusted networks.
func IsIPTrusted(ctx *middlewares.AutheliaCtx, ip net.IP) bool {
	if ip == nil {
		return false
	}

	ctx.Logger.Debugf("IP %s - IgnorePrivateRanges: %t, IsPrivate: %t, IsLoopback: %t",
		ip.String(), ctx.Configuration.AuthenticationBackend.KnownIP.NotifyPrivateRanges, ip.IsPrivate(), ip.IsLoopback())

	if !ctx.Configuration.AuthenticationBackend.KnownIP.NotifyPrivateRanges && (ip.IsPrivate() || ip.IsLoopback()) {
		return true
	}

	for i, network := range ctx.Configuration.AuthenticationBackend.KnownIP.TrustedNetworks {
		if network.Contains(ip) {
			ctx.Logger.Debugf("IP %s is trusted (matches configured network %s)", ip.String(), network.String())
			return true
		}

		ctx.Logger.Debugf("IP %s does not match trusted network %d: %s", ip.String(), i, network.String())
	}

	ctx.Logger.Debugf("IP %s is not trusted", ip.String())

	return false
}

// HandleKnownIPTracking manages IP tracking for user sessions, updating existing IPs or handling new ones.
func HandleKnownIPTracking(ctx *middlewares.AutheliaCtx, userSession *session.UserSession) {
	if !ctx.Configuration.AuthenticationBackend.KnownIP.Enable {
		return
	}

	remoteIP := ctx.RequestCtx.RemoteIP()

	if len(remoteIP) == 0 {
		ctx.Logger.Errorf("Remote IP is invalid, skipping known ip notification for user '%s'", userSession.Username)
		return
	}

	ip := model.NewIP(remoteIP)

	logger := ctx.Logger.WithFields(logrus.Fields{
		"username":   userSession.Username,
		"ip_address": ip.IP.String(),
	})

	if IsIPTrusted(ctx, ip.IP) {
		logger.Debug(logErrSkipTrustedIP)
		return
	}

	ipExists, err := ctx.Providers.StorageProvider.IsIPKnownForUser(ctx, userSession.Username, ip)
	if err != nil {
		logger.WithError(err).Error(logErrCheckKnownIP)
		return
	}

	if ipExists {
		if err = ctx.Providers.StorageProvider.UpdateKnownIP(ctx, userSession.Username, ip); err != nil {
			logger.WithError(err).Error(logErrUpdateKnownIP)
		}
	} else {
		handleNewIP(ctx, userSession, ip)
	}
}

// handleNewIP processes a new IP address by saving it to storage and sending notification email.
func handleNewIP(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, ip model.IP) {
	rawUserAgent := string(ctx.RequestCtx.Request.Header.Peek("User-Agent"))
	userAgent := utils.ParseUserAgent(rawUserAgent)

	logger := ctx.Logger.WithFields(logrus.Fields{
		"username":   userSession.Username,
		"ip_address": ip.IP.String(),
	})

	if err := ctx.Providers.StorageProvider.SaveNewIPForUser(ctx, userSession.Username, ip, *userAgent); err != nil {
		logger.WithError(err).Error(logErrSaveNewKnownIP)
		return
	}

	sendNewIPEmail(ctx, userSession, ip, userAgent, rawUserAgent)
}

// sendNewIPEmail sends an email notification to the user about a login from a new IP address.
func sendNewIPEmail(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, ip model.IP, userAgent *uasurfer.UserAgent, rawUserAgent string) {
	if len(userSession.Emails) == 0 {
		ctx.Logger.Error(fmt.Errorf("user %s has no email address configured", userSession.Username))
		ctx.ReplyOK()

		return
	}

	domain, _ := ctx.GetCookieDomain()

	data := templates.NewEmailNewLoginValues(userSession.DisplayName, domain, ip.String(), userAgent, rawUserAgent, time.Now())

	address, _ := mail.ParseAddress(userSession.Emails[0])

	ctx.Logger.Debugf("Sending an email to user %s (%s) to inform that there is a login from a new ip '%s'.",
		userSession.Username, address.Address, ip.String())

	if err := ctx.Providers.Notifier.Send(ctx, *address, "Login From New IP", ctx.Providers.Templates.GetNewLoginEmailTemplate(), data); err != nil {
		ctx.Logger.Error(err)
		ctx.ReplyOK()

		return
	}
}
