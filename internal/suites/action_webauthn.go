package suites

import (
	"bytes"
	"testing"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/stretchr/testify/require"
)

//nolint:unparam
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
			HasResidentKey:              true,
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

	if !rs.CheckElementExistsLocatedByID(t, page, "one-time-password-method") {
		return
	}

	rs.doWebAuthnMethodMustSelect(t, page)
}

func (rs *RodSession) doWebAuthnMethodMustSelect(t *testing.T, page *rod.Page) {
	rs.ClickElementLocatedByID(t, page, "methods-button")
	rs.ClickElementLocatedByID(t, page, "webauthn-option")
}

func (rs *RodSession) doWebAuthnCredentialMaybeDelete(t *testing.T, page *rod.Page) {
	rs.WaitElementLocatedBySelector(t, page, `#webauthn-credentials-panel[data-loading="false"]`)

	if !rs.CheckElementExistsLocatedByID(t, page, "webauthn-credential-0-delete") {
		return
	}

	rs.doWebAuthnCredentialMustDelete(t, page)
}

func (rs *RodSession) doWebAuthnCredentialMustDelete(t *testing.T, page *rod.Page) {
	rs.ClickElementLocatedByID(t, page, "webauthn-credential-0-delete")

	rs.doMaybeVerifyIdentity(t, page, "#dialog-delete")

	rs.ClickElementLocatedByID(t, page, "dialog-delete")

	rs.verifyNotificationDisplayed(t, page, "Successfully deleted the WebAuthn Credential")

	rs.DeleteWebAuthnAuthenticatorCredentials()
}

func (rs *RodSession) doWebAuthnCredentialRename(t *testing.T, page *rod.Page, description string) {
	rs.ClickElementLocatedByID(t, page, "webauthn-credential-0-edit")

	rs.doMaybeVerifyIdentity(t, page, "#webauthn-credential-description")

	rs.TypeElementLocatedByID(t, page, "webauthn-credential-description", description)

	rs.ClickElementLocatedByID(t, page, "dialog-update")

	rs.verifyNotificationDisplayed(t, page, "Successfully updated the WebAuthn Credential")
}

func (rs *RodSession) doWebAuthnCredentialRegister(t *testing.T, page *rod.Page, description string) {
	rs.doWebAuthnCredentialMaybeDelete(t, page)

	rs.ClickElementLocatedByID(t, page, "webauthn-credential-add")

	rs.doMaybeVerifyIdentity(t, page, "#webauthn-credential-description")

	rs.TypeElementLocatedByID(t, page, "webauthn-credential-description", description)
	rs.ClickElementLocatedByID(t, page, "dialog-next")
	rs.verifyNotificationDisplayed(t, page, "Successfully added the WebAuthn Credential")

	rs.doWebAuthnUpdateCredentials(t, page)

	rs.doOpenSettingsMenuClickClose(t, page)
}

func (rs *RodSession) doWebAuthnCredentialRegisterAfterVisitSettings(t *testing.T, page *rod.Page, description string) {
	rs.doOpenSettings(t, page)
	rs.doOpenSettingsMenuClickTwoFactor(t, page)
	rs.doWebAuthnCredentialRegister(t, page, description)
}
