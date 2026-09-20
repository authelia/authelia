// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/utils"
)

const (
	passwordChangeRequiredAttribute = "pwd_reset"

	passwordChangeRequiredReplaced = "new-password"

	passwordChangeRequiredLDAPAdminDN = "cn=admin,dc=example,dc=com"                        //nolint:gosec // A suite fixture distinguished name, not a credential.
	passwordChangeRequiredLDAPUserDN  = "cn=John Doe (external),ou=users,dc=example,dc=com" //nolint:gosec // A suite fixture distinguished name, not a credential.

	passwordChangeRequiredLDAPPassword = "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/" //nolint:gosec // Suite fixture password hash, not a real credential.
)

// PasswordChangeRequiredBackend puts the user a ChangePasswordScenario runs against into the state which requires a
// password change, and puts them back again afterwards.
//
// It belongs to the suite rather than to the scenario because neither half of that state can be reached through the
// portal: the password is the one the scenario replaces, and the attribute is the one either Authelia or the
// directory server clears. How both are written depends entirely on the authentication backend. Doing it here
// rather than in the fixtures is also what lets the scenario use the same user every other scenario signs in as,
// since the user only carries the attribute while the one test which needs it is running.
type PasswordChangeRequiredBackend interface {
	// Prepare puts the user into the state which requires a password change.
	Prepare(t *testing.T)

	// Reset puts the user back the way the suite found them.
	Reset(t *testing.T)
}

// NewPasswordChangeRequiredFileBackend returns a PasswordChangeRequiredBackend for a suite using the file provider.
func NewPasswordChangeRequiredFileBackend() PasswordChangeRequiredBackend {
	return &PasswordChangeRequiredFileBackend{}
}

// PasswordChangeRequiredFileBackend adjusts the user database of a suite using the file provider.
type PasswordChangeRequiredFileBackend struct {
	path     string
	snapshot []byte
}

// Prepare implements the PasswordChangeRequiredBackend interface.
func (b *PasswordChangeRequiredFileBackend) Prepare(t *testing.T) {
	var err error

	b.path = filepath.Join(os.Getenv("SUITE"), "users.yml")

	b.snapshot, err = os.ReadFile(b.path)

	require.NoError(t, err)

	database := authentication.NewFileUserDatabase(b.path, false, false, map[string]expression.ExtraAttribute{
		passwordChangeRequiredAttribute: schema.AuthenticationBackendExtraAttribute{ValueType: "boolean"},
	})

	require.NoError(t, database.Load())

	details, err := database.GetUserDetails(testUsername)
	require.NoError(t, err)

	if details.Extra == nil {
		details.Extra = map[string]any{}
	}

	details.Extra[passwordChangeRequiredAttribute] = true

	database.SetUserDetails(details.Username, &details)

	since := time.Now().UTC().Add(-time.Second)

	require.NoError(t, database.Save())

	waitForUsersDatabaseReload(t, since)
}

// Reset implements the PasswordChangeRequiredBackend interface.
//
// The database is written back byte for byte rather than edited, as Authelia rewrites the whole file in its own
// canonical form whenever it changes a password, and the fixture is tracked.
func (b *PasswordChangeRequiredFileBackend) Reset(t *testing.T) {
	since := time.Now().UTC().Add(-time.Second)

	require.NoError(t, os.WriteFile(b.path, b.snapshot, 0600))

	waitForUsersDatabaseReload(t, since)
}

// NewPasswordChangeRequiredLDAPBackend returns a PasswordChangeRequiredBackend for a suite which runs OpenLDAP as a
// service of its compose project.
func NewPasswordChangeRequiredLDAPBackend() PasswordChangeRequiredBackend {
	return &PasswordChangeRequiredLDAPBackend{}
}

// NewPasswordChangeRequiredKubernetesLDAPBackend returns a PasswordChangeRequiredBackend for a suite which runs
// OpenLDAP as a pod inside the cluster, where it is reached through the k3d container rather than directly.
func NewPasswordChangeRequiredKubernetesLDAPBackend() PasswordChangeRequiredBackend {
	return &PasswordChangeRequiredLDAPBackend{kubernetes: true}
}

// PasswordChangeRequiredLDAPBackend adjusts the directory entry of a suite using OpenLDAP.
type PasswordChangeRequiredLDAPBackend struct {
	kubernetes bool
}

// Prepare implements the PasswordChangeRequiredBackend interface.
func (b *PasswordChangeRequiredLDAPBackend) Prepare(t *testing.T) {
	b.modify(t, fmt.Sprintf("dn: %s\nchangetype: modify\nreplace: pwdReset\npwdReset: TRUE\n", passwordChangeRequiredLDAPUserDN))
}

// Reset implements the PasswordChangeRequiredBackend interface.
//
// The password and 'pwdReset' are written in separate operations because the ppolicy overlay acts on a password
// modification, and giving it a request which also carries the attribute it maintains invites it to disagree. The
// attribute is replaced with nothing rather than deleted so that a run which ended after the overlay had already
// cleared it does not fail on an attribute which is no longer there.
func (b *PasswordChangeRequiredLDAPBackend) Reset(t *testing.T) {
	b.modify(t, fmt.Sprintf("dn: %s\nchangetype: modify\nreplace: userPassword\nuserPassword: %s\n", passwordChangeRequiredLDAPUserDN, passwordChangeRequiredLDAPPassword))
	b.modify(t, fmt.Sprintf("dn: %s\nchangetype: modify\nreplace: pwdReset\n", passwordChangeRequiredLDAPUserDN))
}

func (b *PasswordChangeRequiredLDAPBackend) modify(t *testing.T, ldif string) {
	args := []string{"exec", "-i"}

	if b.kubernetes {
		args = append(args, fmt.Sprintf("%s-k3d-1", composeProjectName()), "kubectl", "exec", "-i", "-n", "authelia", "deploy/ldap", "--")
	} else {
		args = append(args, fmt.Sprintf("%s-openldap-1", composeProjectName()))
	}

	args = append(args, "ldapmodify", "-x", "-H", "ldap://localhost", "-D", passwordChangeRequiredLDAPAdminDN, "-w", "password")

	cmd := exec.Command("docker", args...)

	cmd.Stdin = strings.NewReader(ldif)

	output, err := cmd.CombinedOutput()

	require.NoError(t, err, string(output))
}

// NewPasswordChangeRequiredActiveDirectoryBackend returns a PasswordChangeRequiredBackend for a suite using Samba
// in its Active Directory domain controller role.
func NewPasswordChangeRequiredActiveDirectoryBackend() PasswordChangeRequiredBackend {
	return &PasswordChangeRequiredActiveDirectoryBackend{}
}

// PasswordChangeRequiredActiveDirectoryBackend adjusts the directory entry of a suite using Active Directory.
type PasswordChangeRequiredActiveDirectoryBackend struct{}

// Prepare implements the PasswordChangeRequiredBackend interface.
//
// Active Directory signals this with a zeroed 'pwdLastSet' rather than an attribute of its own, and zeroing it is
// what '--must-change-at-next-login' does, so the password and the signal are set by the one command.
func (b *PasswordChangeRequiredActiveDirectoryBackend) Prepare(t *testing.T) {
	b.setPassword(t, "--must-change-at-next-login")
}

// Reset implements the PasswordChangeRequiredBackend interface.
//
// Setting the password without that flag stamps 'pwdLastSet' with the time of the reset, which is the same thing
// the directory does for itself when the user changes their own password.
func (b *PasswordChangeRequiredActiveDirectoryBackend) Reset(t *testing.T) {
	b.setPassword(t)
}

func (b *PasswordChangeRequiredActiveDirectoryBackend) setPassword(t *testing.T, args ...string) {
	cmd := exec.Command("docker", "exec", fmt.Sprintf("%s-sambaldap-1", composeProjectName()), //nolint:gosec // Suite fixture command, not attacker controlled.
		"samba-tool", "user", "setpassword", testUsername,
		fmt.Sprintf("--newpassword=%s", testPassword))

	cmd.Args = append(cmd.Args, args...)

	output, err := cmd.CombinedOutput()

	require.NoError(t, err, string(output))
}

func waitForUsersDatabaseReload(t *testing.T, since time.Time) {
	name := fmt.Sprintf("%s-authelia-backend-1", composeProjectName())

	require.NoError(t, utils.CheckUntil(500*time.Millisecond, 30*time.Second, func() (bool, error) {
		logs, err := exec.Command("docker", "logs", "--since", since.Format(time.RFC3339), name).CombinedOutput() //nolint:gosec // Suite fixture command, not attacker controlled.
		if err != nil {
			return false, nil //nolint:nilerr
		}

		return strings.Contains(string(logs), "Reloaded successfully"), nil
	}))
}
