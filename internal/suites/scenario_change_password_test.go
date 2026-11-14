// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"
)

type ChangePasswordScenario struct {
	*RodSuite

	backend PasswordChangeRequiredBackend
}

func NewChangePasswordScenario(backend PasswordChangeRequiredBackend) *ChangePasswordScenario {
	return &ChangePasswordScenario{RodSuite: NewRodSuite(""), backend: backend}
}

func (s *ChangePasswordScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *ChangePasswordScenario) TearDownSuite() {
	err := s.Stop()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *ChangePasswordScenario) SetupTest() {
	s.doSetupTest(HomeBaseURL)
}

func (s *ChangePasswordScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

func (s *ChangePasswordScenario) TestShouldChangePassword() {
	testCases := []struct {
		name         string
		username     string
		oldPassword  string
		newPassword  string
		notification string
	}{
		{"NewPassword", testUsername, testPassword, "password1", "Password changed successfully"},
		{"OriginalPassword", testUsername, "password1", testPassword, "Password changed successfully"},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)

			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			s.doLoginOneFactor(s.T(), s.Context(ctx), tc.username, tc.oldPassword, false, BaseDomain, "")
			s.doOpenSettings(s.T(), s.Context(ctx))
			s.doOpenSettingsMenuClickSecurity(s.T(), s.Context(ctx))

			s.doChangePassword(s.T(), s.Context(ctx), tc.oldPassword, tc.newPassword, tc.newPassword, tc.notification)
			s.doLogout(s.T(), s.Context(ctx))
		})
	}
}

func (s *ChangePasswordScenario) TestShouldNotChangePasswordToExistingPassword() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), testUsername, testPassword, false, BaseDomain, "")
	s.doOpenSettings(s.T(), s.Context(ctx))
	s.doOpenSettingsMenuClickSecurity(s.T(), s.Context(ctx))

	s.doChangePassword(s.T(), s.Context(ctx), testPassword, testPassword, testPassword, "Your supplied password does not meet the password policy requirements")

	s.doLogout(s.T(), s.Context(ctx))
}

func (s *ChangePasswordScenario) TestShouldNotChangePasswordWithIncorrectOldPassword() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), testUsername, testPassword, false, BaseDomain, "")

	s.doOpenSettings(s.T(), s.Context(ctx))

	s.doOpenSettingsMenuClickSecurity(s.T(), s.Context(ctx))

	s.doChangePassword(s.T(), s.Context(ctx), "wrong_password", "new_password", "new_password", "Incorrect password")

	s.doLogout(s.T(), s.Context(ctx))
}

func (s *ChangePasswordScenario) TestShouldNotChangePasswordNewPasswordsMustMatch() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), testUsername, testPassword, false, BaseDomain, "")

	s.doOpenSettings(s.T(), s.Context(ctx))

	s.doOpenSettingsMenuClickSecurity(s.T(), s.Context(ctx))

	s.doChangePassword(s.T(), s.Context(ctx), testPassword, "my_new_password", "new_password", "Passwords do not match")

	s.doLogout(s.T(), s.Context(ctx))
}

func (s *ChangePasswordScenario) TestShouldRequireAPasswordChangeBeforeAuthenticating() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.backend.Prepare(s.T())

	defer s.backend.Reset(s.T())

	targetURL := fmt.Sprintf("%s/secret.html", SingleFactorBaseURL)

	s.doLoginOneFactor(s.T(), s.Context(ctx), testUsername, testPassword, false, BaseDomain, targetURL)

	s.verifyIsPasswordChangeRequiredPage(s.T(), s.Context(ctx))

	// The held session carries no authentication level, so a resource which only requires one factor is still out
	// of reach with the password the administrator issued.
	s.doVisit(s.T(), s.Context(ctx), targetURL)
	s.verifyIsPasswordChangeRequiredPage(s.T(), s.Context(ctx))

	s.doChangeRequiredPassword(s.T(), s.Context(ctx), testPassword, passwordChangeRequiredReplaced, "another-password")
	s.verifyNotificationDisplayed(s.T(), s.Context(ctx), "Passwords do not match")
	s.verifyIsPasswordChangeRequiredPage(s.T(), s.Context(ctx))

	s.doChangeRequiredPassword(s.T(), s.Context(ctx), testPassword, passwordChangeRequiredReplaced, passwordChangeRequiredReplaced)

	// The session which held the user is destroyed by the change, so they arrive back at the first factor page and
	// sign in with the password they just set.
	s.verifyIsFirstFactorPage(s.T(), s.Context(ctx))

	s.doLoginOneFactor(s.T(), s.Context(ctx), testUsername, passwordChangeRequiredReplaced, false, BaseDomain, targetURL)
	s.verifySecretAuthorized(s.T(), s.Context(ctx))
}
