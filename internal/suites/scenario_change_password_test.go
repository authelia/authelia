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
	err := s.RodSession.Stop()

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
		name        string
		username    string
		oldPassword string
		newPassword string
	}{
		{"case1", "john", "password", "password1"},
		{"case1", "john", "password1", "password"},
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

			s.doChangePassword(s.T(), s.Context(ctx), tc.oldPassword, tc.newPassword, tc.newPassword)
			s.doLogout(s.T(), s.Context(ctx))
		})
	}
}

func (s *ChangePasswordScenario) TestCannotChangePasswordToExistingPassword() {
	testCases := []struct {
		testName    string
		username    string
		oldPassword string
	}{
		{"case1", "john", "password"},
	}

	for _, tc := range testCases {
		s.T().Run(tc.testName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()
			s.doLoginOneFactor(s.T(), s.Context(ctx), tc.username, tc.oldPassword, false, BaseDomain, "")

			s.doOpenSettings(s.T(), s.Context(ctx))

			s.doOpenSettingsMenuClickSecurity(s.T(), s.Context(ctx))

			s.doMustChangePasswordExistingPassword(s.T(), s.Context(ctx), tc.oldPassword, tc.oldPassword)

			s.doLogout(s.T(), s.Context(ctx))
		})
	}
}

func (s *ChangePasswordScenario) TestCannotChangePasswordWithIncorrectOldPassword() {
	testCases := []struct {
		testName         string
		username         string
		oldPassword      string
		wrongOldPassword string
		newPassword      string
	}{
		{"case1", "john", "password", "wrong_password", "new_password"},
	}

	for _, tc := range testCases {
		s.T().Run(tc.testName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()
			s.doLoginOneFactor(s.T(), s.Context(ctx), tc.username, tc.oldPassword, false, BaseDomain, "")

			s.doOpenSettings(s.T(), s.Context(ctx))

			s.doOpenSettingsMenuClickSecurity(s.T(), s.Context(ctx))

			s.doMustChangePasswordWrongExistingPassword(s.T(), s.Context(ctx), tc.wrongOldPassword, tc.newPassword)

			s.doLogout(s.T(), s.Context(ctx))
		})
	}
}

func (s *ChangePasswordScenario) TestNewPasswordsMustMatch() {
	testCases := []struct {
		testName     string
		username     string
		oldPassword  string
		newPassword1 string
		newPassword2 string
	}{
		{"case1", "john", "password", "my_new_password", "new_password"},
	}

	for _, tc := range testCases {
		s.T().Run(tc.testName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()
			s.doLoginOneFactor(s.T(), s.Context(ctx), tc.username, tc.oldPassword, false, BaseDomain, "")

			s.doOpenSettings(s.T(), s.Context(ctx))

			s.doOpenSettingsMenuClickSecurity(s.T(), s.Context(ctx))

			s.doMustChangePasswordMustMatch(s.T(), s.Context(ctx), tc.oldPassword, tc.newPassword1, tc.newPassword2)

			s.doLogout(s.T(), s.Context(ctx))
		})
	}
}
