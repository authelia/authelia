package session

import (
	"maps"
	"slices"
)

func (s UserSession) deepCopy() (session UserSession) {
	session = s

	session.AuthenticationMethodRefs.Extra = slices.Clone(s.AuthenticationMethodRefs.Extra)
	session.WebAuthn = s.WebAuthn.deepCopy()
	session.Elevations.User = s.Elevations.User.deepCopy()

	if s.PasswordResetUsername != nil {
		username := *s.PasswordResetUsername

		session.PasswordResetUsername = &username
	}

	if s.TOTP != nil {
		totp := *s.TOTP

		session.TOTP = &totp
	}

	return session
}

func (w *WebAuthn) deepCopy() (webAuthn *WebAuthn) {
	if w == nil {
		return nil
	}

	value := *w

	if w.SessionData != nil {
		data := *w.SessionData

		data.UserID = slices.Clone(w.UserID)
		data.CredParams = slices.Clone(w.CredParams)
		data.Extensions = maps.Clone(w.Extensions)

		if w.AllowedCredentialIDs != nil {
			data.AllowedCredentialIDs = make([][]byte, len(w.AllowedCredentialIDs))

			for i, id := range w.AllowedCredentialIDs {
				data.AllowedCredentialIDs[i] = slices.Clone(id)
			}
		}

		value.SessionData = &data
	}

	return &value
}

func (e *Elevation) deepCopy() (elevation *Elevation) {
	if e == nil {
		return nil
	}

	value := *e

	value.RemoteIP = slices.Clone(e.RemoteIP)

	return &value
}
