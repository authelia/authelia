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
	s.Page.MustElementR(rowSelector(userManagementTableID, "user-row-"+uiTestUserUsername), uiTestUserDisplayName2)

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
