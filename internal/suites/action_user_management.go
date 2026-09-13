package suites

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
)

// The user management UI actions locate elements exclusively by the stable ids and row classes documented in
// the user management frontend (user-management-add, user-row-<username>-edit, new-user-<field>, etc.). Those
// hooks are part of the UI contract and must survive component library changes so this scenario keeps passing.

const (
	userManagementTableID  = "user-management-table"
	groupManagementTableID = "group-management-table"
	rowWaitTimeout         = 10 * time.Second
)

func (rs *RodSession) doOpenSettingsMenuClickUsers(t *testing.T, page *rod.Page) {
	rs.doOpenSettingsMenu(t, page)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "settings-menu-users").Click("left", 1))

	require.NoError(t, page.WaitStable(time.Millisecond*100))
}

func (rs *RodSession) doOpenSettingsMenuClickGroups(t *testing.T, page *rod.Page) {
	rs.doOpenSettingsMenu(t, page)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "settings-menu-groups").Click("left", 1))

	require.NoError(t, page.WaitStable(time.Millisecond*100))
}

// verifySettingsMenuItemsVisible opens the settings drawer and asserts the presence or absence of the given
// nav item keynames (e.g. "users", "groups").
func (rs *RodSession) verifySettingsMenuItemsVisible(t *testing.T, page *rod.Page, present []string, absent []string) {
	rs.doOpenSettingsMenu(t, page)

	require.NoError(t, page.WaitStable(time.Millisecond*100))

	for _, keyname := range present {
		require.True(t, rs.CheckElementExistsLocatedByID(t, page, "settings-menu-"+keyname), "expected settings menu item %q to be present", keyname)
	}

	for _, keyname := range absent {
		require.False(t, rs.CheckElementExistsLocatedByID(t, page, "settings-menu-"+keyname), "expected settings menu item %q to be absent", keyname)
	}
}

// doFillTextInputByID replaces the content of a text input located by id.
func (rs *RodSession) doFillTextInputByID(t *testing.T, page *rod.Page, id, value string) {
	element := rs.WaitElementLocatedByID(t, page, id)

	require.NoError(t, element.MustSelectAllText().Input(value))
}

// doFillTextInputByIDIfPresent fills the input only when it exists, which lets the flows cope with
// backends that mark different attributes as required.
func (rs *RodSession) doFillTextInputByIDIfPresent(t *testing.T, page *rod.Page, id, value string) {
	if !rs.CheckElementExistsLocatedByID(t, page, id) {
		return
	}

	rs.doFillTextInputByID(t, page, id, value)
}

// doClickIfPresent clicks the element when it exists.
func (rs *RodSession) doClickIfPresent(t *testing.T, page *rod.Page, id string) {
	if !rs.CheckElementExistsLocatedByID(t, page, id) {
		return
	}

	require.NoError(t, rs.WaitElementLocatedByID(t, page, id).Click("left", 1))
	require.NoError(t, page.WaitStable(time.Millisecond*100))
}

// doScrollTableRight scrolls a horizontally virtualised table to its right edge so the trailing actions column is
// rendered. It is a no-op for tables without an inner scroll container.
func (rs *RodSession) doScrollTableRight(t *testing.T, page *rod.Page, tableID string) {
	_, err := page.Eval(`(id) => {
		const root = document.getElementById(id);
		if (!root) return;
		const scroller = root.querySelector('.MuiDataGrid-virtualScroller') || root.querySelector('[data-slot="table-container"]') || root;
		scroller.scrollLeft = scroller.scrollWidth;
	}`, tableID)

	require.NoError(t, err)
	require.NoError(t, page.WaitStable(time.Millisecond*100))
}

func rowSelector(tableID, rowClass string) string {
	return fmt.Sprintf("#%s .%s", tableID, rowClass)
}

// waitRowAbsent polls until no element matches the row selector or the timeout elapses.
func (rs *RodSession) waitRowAbsent(t *testing.T, page *rod.Page, selector string) {
	deadline := time.Now().Add(rowWaitTimeout)

	for time.Now().Before(deadline) {
		if !rs.CheckElementExistsLocatedBySelector(t, page, selector) {
			return
		}

		time.Sleep(200 * time.Millisecond)
	}

	require.Failf(t, "row still present", "element matching %q is still present after %s", selector, rowWaitTimeout)
}

func (rs *RodSession) verifyUserRowExists(t *testing.T, page *rod.Page, username string) {
	rs.WaitElementLocatedBySelector(t, page, rowSelector(userManagementTableID, "user-row-"+username))
}

func (rs *RodSession) verifyUserRowAbsent(t *testing.T, page *rod.Page, username string) {
	rs.waitRowAbsent(t, page, rowSelector(userManagementTableID, "user-row-"+username))
}

func (rs *RodSession) verifyGroupRowExists(t *testing.T, page *rod.Page, name string) {
	rs.WaitElementLocatedBySelector(t, page, rowSelector(groupManagementTableID, "group-row-"+name))
}

func (rs *RodSession) verifyGroupRowAbsent(t *testing.T, page *rod.Page, name string) {
	rs.waitRowAbsent(t, page, rowSelector(groupManagementTableID, "group-row-"+name))
}

// doUserManagementAddUser creates a user through the new user dialog. Fields which a backend does not expose or
// require are skipped.
func (rs *RodSession) doUserManagementAddUser(t *testing.T, page *rod.Page, username, password, email, displayName, familyName string) {
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "user-management-add").Click("left", 1))

	rs.WaitElementLocatedByID(t, page, "new-user-dialog")
	rs.WaitElementLocatedByID(t, page, "new-user-username")

	rs.doFillTextInputByID(t, page, "new-user-username", username)
	rs.doFillTextInputByID(t, page, "new-user-password", password)
	rs.doFillTextInputByIDIfPresent(t, page, "new-user-mail", email)

	rs.doClickIfPresent(t, page, "new-user-toggle-additional")

	rs.doFillTextInputByIDIfPresent(t, page, "new-user-family_name", familyName)
	rs.doFillTextInputByIDIfPresent(t, page, "new-user-display_name", displayName)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "new-user-submit").Click("left", 1))

	rs.verifyNotificationDisplayed(t, page, "User created successfully")
}

// doUserManagementEditDisplayName edits the display name of a user through the edit user dialog.
func (rs *RodSession) doUserManagementEditDisplayName(t *testing.T, page *rod.Page, username, displayName string) {
	rs.doScrollTableRight(t, page, userManagementTableID)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, fmt.Sprintf("user-row-%s-edit", username)).Click("left", 1))

	rs.WaitElementLocatedByID(t, page, "edit-user-dialog")
	rs.WaitElementLocatedByID(t, page, "edit-user-username")

	if !rs.CheckElementExistsLocatedByID(t, page, "edit-user-display_name") {
		rs.doClickIfPresent(t, page, "edit-user-toggle-additional")
	}

	rs.doFillTextInputByID(t, page, "edit-user-display_name", displayName)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "edit-user-submit").Click("left", 1))

	rs.verifyNotificationDisplayed(t, page, "User modified successfully")
}

// doUserManagementSetPassword sets a user's password through the row menu and set password dialog.
func (rs *RodSession) doUserManagementSetPassword(t *testing.T, page *rod.Page, username, password string) {
	rs.doScrollTableRight(t, page, userManagementTableID)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, fmt.Sprintf("user-row-%s-more", username)).Click("left", 1))
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "user-menu-change-password").Click("left", 1))

	rs.WaitElementLocatedByID(t, page, "set-password-dialog")

	rs.doFillTextInputByID(t, page, "set-password-password", password)
	rs.doFillTextInputByID(t, page, "set-password-confirm", password)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "set-password-submit").Click("left", 1))

	rs.verifyNotificationDisplayed(t, page, "Password updated successfully")
}

// doUserManagementDeleteUser deletes a user through the type-to-confirm dialog.
func (rs *RodSession) doUserManagementDeleteUser(t *testing.T, page *rod.Page, username string) {
	rs.doScrollTableRight(t, page, userManagementTableID)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, fmt.Sprintf("user-row-%s-delete", username)).Click("left", 1))

	rs.WaitElementLocatedByID(t, page, "verify-delete-user-dialog")

	confirm := rs.WaitElementLocatedByID(t, page, "verify-delete-user-confirm")

	require.True(t, confirm.MustProperty("disabled").Bool(), "delete must be disabled before the username is typed")

	rs.doFillTextInputByID(t, page, "verify-delete-user-input", username)

	require.NoError(t, confirm.Click("left", 1))

	rs.verifyNotificationDisplayed(t, page, "User deleted successfully")
}

// doGroupManagementAddGroup creates a group through the new group dialog.
func (rs *RodSession) doGroupManagementAddGroup(t *testing.T, page *rod.Page, name string) {
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "group-management-add").Click("left", 1))

	rs.WaitElementLocatedByID(t, page, "new-group-dialog")

	rs.doFillTextInputByID(t, page, "new-group-name", name)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "new-group-submit").Click("left", 1))

	rs.verifyNotificationDisplayed(t, page, "Group created successfully")
}

// doGroupManagementDeleteGroup deletes a group through the type-to-confirm dialog.
func (rs *RodSession) doGroupManagementDeleteGroup(t *testing.T, page *rod.Page, name string) {
	rs.doScrollTableRight(t, page, groupManagementTableID)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, fmt.Sprintf("group-row-%s-delete", name)).Click("left", 1))

	rs.WaitElementLocatedByID(t, page, "verify-delete-group-dialog")

	rs.doFillTextInputByID(t, page, "verify-delete-group-input", name)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "verify-delete-group-confirm").Click("left", 1))

	rs.verifyNotificationDisplayed(t, page, "Group deleted successfully")
}
