package suites

import (
	"bytes"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/stretchr/testify/require"
)

func (rs *RodSession) doWebAuthnInitialize(t *testing.T, page *rod.Page, enableUI bool) {
	rs.doWebAuthnEnable(t, page, enableUI)

	rs.doWebAuthnAddVirtualAuthenticator(t, page)
}

func (rs *RodSession) doWebAuthnRestoreCredentials(t *testing.T, page *rod.Page) {
	authenticatorID := rs.GetWebAuthnAuthenticatorID()

	credentials := rs.GetWebAuthnCredentials()

	ecredentials := rs.doWebAuthnGetCredentials(t, page)

	if len(credentials) != 0 {
	outer:
		for _, credential := range credentials {
			for _, existing := range ecredentials.Credentials {
				if bytes.Equal(existing.CredentialID, credential.CredentialID) {
					continue outer
				}
			}

			rs.doWebAuthnAddCredential(t, page, authenticatorID, credential)
		}
	}
}

func (rs *RodSession) doWebAuthnEnable(t *testing.T, client proto.Client, enableUI bool) {
	require.NoError(t, proto.WebAuthnEnable{EnableUI: enableUI}.Call(client))
}

func (rs *RodSession) doWebAuthnAddVirtualAuthenticator(t *testing.T, page *rod.Page) (result *proto.WebAuthnAddVirtualAuthenticatorResult) {
	result, err := proto.WebAuthnAddVirtualAuthenticator{
		Options: &proto.WebAuthnVirtualAuthenticatorOptions{
			Protocol:                    proto.WebAuthnAuthenticatorProtocolCtap2,
			Ctap2Version:                proto.WebAuthnCtap2VersionCtap21,
			Transport:                   proto.WebAuthnAuthenticatorTransportNfc,
			HasUserVerification:         true,
			AutomaticPresenceSimulation: true,
			IsUserVerified:              true,
		},
	}.Call(page)

	require.NoError(t, err)

	rs.SetWebAuthnAuthenticatorID(result.AuthenticatorID)

	return result
}

func (rs *RodSession) doWebAuthnAddCredential(t *testing.T, page *rod.Page, authenticatorID proto.WebAuthnAuthenticatorID, credential *proto.WebAuthnCredential) {
	require.NoError(t, proto.WebAuthnAddCredential{AuthenticatorID: authenticatorID, Credential: credential}.Call(page))
}

func (rs *RodSession) doWebAuthnGetCredentials(t *testing.T, page *rod.Page) *proto.WebAuthnGetCredentialsResult {
	result, err := proto.WebAuthnGetCredentials{AuthenticatorID: rs.GetWebAuthnAuthenticatorID()}.Call(page)
	require.NoError(t, err)

	return result
}

func (rs *RodSession) doWebAuthnUpdateCredentials(t *testing.T, page *rod.Page) {
	result := rs.doWebAuthnGetCredentials(t, page)

	rs.SetWebAuthnAuthenticatorCredentials(result.Credentials...)
}

func (rs *RodSession) doWebAuthnMethodMaybeSelect(t *testing.T, page *rod.Page) {
	_ = rs.WaitElementLocatedByID(t, page, "second-factor-stage")

	has, _, err := page.Has("#one-time-password-method")
	require.NoError(t, err)

	if !has {
		return
	}

	rs.doWebAuthnMethodMustSelect(t, page)
}

func (rs *RodSession) doWebAuthnMethodMustSelect(t *testing.T, page *rod.Page) {
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "methods-button").Click("left", 1))
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "webauthn-option").Click("left", 1))
}

func (rs *RodSession) doWebAuthnCredentialMaybeDelete(t *testing.T, page *rod.Page) {
	require.NoError(t, page.WaitStable(time.Millisecond*100))

	has, _, err := page.Has("#webauthn-credential-0-delete")
	require.NoError(t, err)

	if !has {
		return
	}

	rs.doWebAuthnCredentialMustDelete(t, page)
}

func (rs *RodSession) doWebAuthnCredentialMustDelete(t *testing.T, page *rod.Page) {
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "webauthn-credential-0-delete").Click("left", 1))

	rs.doMaybeVerifyIdentity(t, page)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "dialog-delete").Click("left", 1))

	rs.verifyNotificationDisplayed(t, page, "Successfully deleted the WebAuthn Credential")

	rs.DeleteWebAuthnAuthenticatorCredentials()
}

func (rs *RodSession) doWebAuthnCredentialRename(t *testing.T, page *rod.Page, description string) {
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "webauthn-credential-0-edit").Click("left", 1))

	rs.doMaybeVerifyIdentity(t, page)

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "webauthn-credential-description").Type(rs.toInputs(description)...))

	require.NoError(t, rs.WaitElementLocatedByID(t, page, "dialog-update").Click("left", 1))

	rs.verifyNotificationDisplayed(t, page, "Successfully updated the WebAuthn Credential")
}

func (rs *RodSession) doWebAuthnCredentialRegister(t *testing.T, page *rod.Page, description string) {
	rs.doWebAuthnCredentialMaybeDelete(t, page)

	elementAdd := rs.WaitElementLocatedByID(t, page, "webauthn-credential-add")

	require.NoError(t, elementAdd.Click("left", 1))

	rs.doMaybeVerifyIdentity(t, page)

	elementDescription := rs.WaitElementLocatedByID(t, page, "webauthn-credential-description")

	require.NoError(t, elementDescription.Type(rs.toInputs(description)...))
	require.NoError(t, rs.WaitElementLocatedByID(t, page, "dialog-next").Click("left", 1))
	rs.verifyNotificationDisplayed(t, page, "Successfully added the WebAuthn Credential")

	rs.doWebAuthnUpdateCredentials(t, page)

	require.NoError(t, page.WaitStable(time.Millisecond*50))
	rs.doHoverAllMuiTooltip(t, page)
	require.NoError(t, page.WaitStable(time.Millisecond*50))

	rs.doOpenSettingsMenuClickClose(t, page)
}

func (rs *RodSession) doWebAuthnCredentialRegisterAfterVisitSettings(t *testing.T, page *rod.Page, description string) {
	rs.doOpenSettings(t, page)
	rs.doOpenSettingsMenuClickTwoFactor(t, page)
	rs.doWebAuthnCredentialRegister(t, page, description)
}
