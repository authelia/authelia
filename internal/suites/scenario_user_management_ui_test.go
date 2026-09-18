package suites

import (
	"context"
	"log"
	"time"
)

// UserManagementUIScenario exercises the administrative user and group management settings pages through the
// browser. It requires a suite with administration.enabled, enable_user_management and an admin user (john) in the
// admin group, plus a directory backend for the group management page.
type UserManagementUIScenario struct {
	*RodSuite
}

func NewUserManagementUIScenario() *UserManagementUIScenario {
	return &UserManagementUIScenario{RodSuite: NewRodSuite("")}
}

func (s *UserManagementUIScenario) SetupSuite() {
	browser, err := NewRodSession(RodSessionWithCredentials(s))
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser
}

func (s *UserManagementUIScenario) TearDownSuite() {
	err := s.Stop()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *UserManagementUIScenario) SetupTest() {
	s.Page = s.doCreateTab(s.T(), HomeBaseURL)
	s.verifyIsHome(s.T(), s.Page)
}

func (s *UserManagementUIScenario) TearDownTest() {
	s.collectCoverage(s.Page)
	s.MustClose()
}

const (
	uiTestUserUsername     = "uitestuser"
	uiTestUserPassword     = "UiTestPassw0rd!"
	uiTestUserNewPassword  = "UiTestNewPassw0rd!"
	uiTestUserEmail        = "uitestuser@example.com"
	uiTestUserDisplayName  = "UI Test User"
	uiTestUserFamilyName   = "Tester"
	uiTestUserDisplayName2 = "UI Test User Renamed"
	uiTestGroupName        = "uitestgroup"

	uiTestResetUserUsername    = "uitestresetuser"
	uiTestResetUserPassword    = "UiTestResetPassw0rd!"
	uiTestResetUserNewPassword = "UiTestResetNewPassw0rd!"
	uiTestResetUserEmail       = "uitestresetuser@example.com"
	uiTestResetUserDisplayName = "UI Test Reset User"
	uiTestResetUserFamilyName  = "Resetter"
)

func (s *UserManagementUIScenario) TestShouldHideManagementItemsFromNonAdmin() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), nonAdminUsername, nonAdminPassword, false, BaseDomain, "")
	s.doOpenSettings(s.T(), s.Context(ctx))
	s.verifySettingsMenuItemsVisible(s.T(), s.Context(ctx), []string{"security"}, []string{"users", "groups"})
	s.doLogout(s.T(), s.Context(ctx))
}

func (s *UserManagementUIScenario) TestShouldShowManagementItemsToAdmin() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), adminUsername, adminPassword, false, BaseDomain, "")
	s.doOpenSettings(s.T(), s.Context(ctx))
	s.verifySettingsMenuItemsVisible(s.T(), s.Context(ctx), []string{"security", "users", "groups"}, nil)
	s.doLogout(s.T(), s.Context(ctx))
}

func (s *UserManagementUIScenario) TestShouldManageUserLifecycle() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	// Create.
	s.doLoginOneFactor(s.T(), s.Context(ctx), adminUsername, adminPassword, false, BaseDomain, "")
	s.doOpenSettings(s.T(), s.Context(ctx))
	s.doOpenSettingsMenuClickUsers(s.T(), s.Context(ctx))
	s.verifyUserRowExists(s.T(), s.Context(ctx), adminUsername)
	s.doUserManagementAddUser(s.T(), s.Context(ctx), uiTestUserUsername, uiTestUserPassword, uiTestUserEmail, uiTestUserDisplayName, uiTestUserFamilyName)
	s.verifyUserRowExists(s.T(), s.Context(ctx), uiTestUserUsername)

	// Edit.
	s.doUserManagementEditDisplayName(s.T(), s.Context(ctx), uiTestUserUsername, uiTestUserDisplayName2)
	s.verifyUserRowExists(s.T(), s.Context(ctx), uiTestUserUsername)
	s.MustElementR(rowSelector(userManagementTableID, "user-row-"+uiTestUserUsername), uiTestUserDisplayName2)

	// Set password.
	s.doUserManagementSetPassword(s.T(), s.Context(ctx), uiTestUserUsername, uiTestUserNewPassword)
	s.doLogout(s.T(), s.Context(ctx))

	// The new user can authenticate with the password set by the admin.
	s.doLoginOneFactor(s.T(), s.Context(ctx), uiTestUserUsername, uiTestUserNewPassword, false, BaseDomain, "")
	s.WaitElementLocatedByID(s.T(), s.Context(ctx), "account-menu")
	s.doLogout(s.T(), s.Context(ctx))

	// Delete.
	s.doLoginOneFactor(s.T(), s.Context(ctx), adminUsername, adminPassword, false, BaseDomain, "")
	s.doOpenSettings(s.T(), s.Context(ctx))
	s.doOpenSettingsMenuClickUsers(s.T(), s.Context(ctx))
	s.verifyUserRowExists(s.T(), s.Context(ctx), uiTestUserUsername)
	s.doUserManagementDeleteUser(s.T(), s.Context(ctx), uiTestUserUsername)
	s.verifyUserRowAbsent(s.T(), s.Context(ctx), uiTestUserUsername)
	s.doLogout(s.T(), s.Context(ctx))
}

// TestShouldSendAndCompletePasswordResetEmail verifies that the admin-triggered "Send Password Reset Email"
// action results in a real email being delivered to the target user, and that following the link in that email
// through to completion lets the user authenticate with their new password. This is kept as its own test (rather
// than folded into TestShouldManageUserLifecycle) because it exercises a distinct user and a full logout/email/
// login round trip, and keeping it isolated makes the failure mode a lot easier to diagnose.
func (s *UserManagementUIScenario) TestShouldSendAndCompletePasswordResetEmail() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	// Create the target user as admin.
	s.doLoginOneFactor(s.T(), s.Context(ctx), adminUsername, adminPassword, false, BaseDomain, "")
	s.doOpenSettings(s.T(), s.Context(ctx))
	s.doOpenSettingsMenuClickUsers(s.T(), s.Context(ctx))
	s.doUserManagementAddUser(s.T(), s.Context(ctx), uiTestResetUserUsername, uiTestResetUserPassword, uiTestResetUserEmail, uiTestResetUserDisplayName, uiTestResetUserFamilyName)
	s.verifyUserRowExists(s.T(), s.Context(ctx), uiTestResetUserUsername)

	// Trigger the password reset email as admin.
	s.doUserManagementSendPasswordResetEmail(s.T(), s.Context(ctx), uiTestResetUserUsername)

	// Verify a real email was actually sent to the target user, not just that a service function was called.
	message := doGetLastEmailMessageWithSubject(s.T(), "[Authelia] Reset your password")

	recipients := make([]string, 0, len(message.To))

	for _, to := range message.To {
		recipients = append(recipients, to.Address)
	}

	s.Assert().Contains(recipients, uiTestResetUserEmail)

	s.doLogout(s.T(), s.Context(ctx))

	// Prove the email is fully functional end to end: follow the link, set a new password, and log in with it.
	s.doSuccessfullyCompletePasswordReset(s.T(), s.Context(ctx), uiTestResetUserNewPassword, uiTestResetUserNewPassword)
	s.doLoginOneFactor(s.T(), s.Context(ctx), uiTestResetUserUsername, uiTestResetUserNewPassword, false, BaseDomain, "")
	s.WaitElementLocatedByID(s.T(), s.Context(ctx), "account-menu")
	s.doLogout(s.T(), s.Context(ctx))

	// Clean up.
	s.doLoginOneFactor(s.T(), s.Context(ctx), adminUsername, adminPassword, false, BaseDomain, "")
	s.doOpenSettings(s.T(), s.Context(ctx))
	s.doOpenSettingsMenuClickUsers(s.T(), s.Context(ctx))
	s.verifyUserRowExists(s.T(), s.Context(ctx), uiTestResetUserUsername)
	s.doUserManagementDeleteUser(s.T(), s.Context(ctx), uiTestResetUserUsername)
	s.verifyUserRowAbsent(s.T(), s.Context(ctx), uiTestResetUserUsername)
	s.doLogout(s.T(), s.Context(ctx))
}

func (s *UserManagementUIScenario) TestShouldManageGroupLifecycle() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	s.doLoginOneFactor(s.T(), s.Context(ctx), adminUsername, adminPassword, false, BaseDomain, "")
	s.doOpenSettings(s.T(), s.Context(ctx))
	s.doOpenSettingsMenuClickGroups(s.T(), s.Context(ctx))
	s.verifyGroupRowExists(s.T(), s.Context(ctx), "dev")
	s.doGroupManagementAddGroup(s.T(), s.Context(ctx), uiTestGroupName)
	s.verifyGroupRowExists(s.T(), s.Context(ctx), uiTestGroupName)
	s.doGroupManagementDeleteGroup(s.T(), s.Context(ctx), uiTestGroupName)
	s.verifyGroupRowAbsent(s.T(), s.Context(ctx), uiTestGroupName)
	s.doLogout(s.T(), s.Context(ctx))
}
