package handlers

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestHandleAuthnCookieValidateActivity(t *testing.T) {
	testCases := []struct {
		Name           string
		Inactivity     time.Duration
		Anonymous      bool
		KeepMeLoggedIn bool
		Elapsed        time.Duration
		Expected       bool
	}{
		{"ShouldNotSaveWhenActivityIsCurrent", time.Minute * 5, false, false, 0, false},
		{"ShouldNotSaveWhenActivityIsRecent", time.Minute * 5, false, false, time.Second * 5, false},
		{"ShouldSaveWhenActivityIsStale", time.Minute * 5, false, false, time.Minute, true},
		{"ShouldNotSaveWhenInactivityDisabled", 0, false, false, time.Hour, false},
		{"ShouldNotSaveWhenAnonymous", time.Minute * 5, true, false, time.Hour, false},
		{"ShouldNotSaveWhenKeepMeLoggedIn", time.Minute * 5, false, true, time.Hour, false},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.Session.Cookies[0].Inactivity = tc.Inactivity
			mock.ResetSessionProvider()

			manager, err := mock.Ctx.GetSessionManagerByTargetURI(&url.URL{Scheme: "https", Host: "app.example.com"})
			require.NoError(t, err)

			require.Equal(t, tc.Inactivity, manager.GetSessionConfig().Inactivity)

			now := mock.Ctx.GetClock().Now()

			userSession := manager.NewDefaultUserSession()

			if !tc.Anonymous {
				userSession = session.NewUserSession("john")
				userSession.CookieDomain = "example.com"
				userSession.SetOneFactorPassword(now, tc.KeepMeLoggedIn)
			}

			userSession.LastActivity = now.Add(-tc.Elapsed).Unix()

			recorded := userSession.LastActivity

			modified, invalid := handleAuthnCookieValidate(mock.Ctx, manager, &userSession)

			require.False(t, invalid)
			assert.Equal(t, tc.Expected, modified)

			if tc.Expected {
				assert.Equal(t, now.Unix(), userSession.LastActivity)
			} else {
				assert.Equal(t, recorded, userSession.LastActivity)
			}
		})
	}
}
