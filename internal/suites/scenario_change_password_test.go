package suites

import (
	"context"
	"log"
	"testing"
	"time"
)

type ChangePasswordScenario struct {
	*RodSuite
}

func NewChangePasswordScenario() *ChangePasswordScenario {
	return &ChangePasswordScenario{RodSuite: NewRodSuite("")}
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
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
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
