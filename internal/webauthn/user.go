package webauthn

import (
	"encoding/binary"

	"github.com/duo-labs/webauthn/webauthn"
)

type User struct {
	ID          int
	Name        string                `json:"name"`
	DisplayName string                `json:"display_name"`
	Icon        string                `json:"icon,omitempty"`
	Credentials []webauthn.Credential `json:"credentials,omitempty"`
}

func (u User) WebAuthnID() []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(buf, int64(u.ID))

	return buf
}

func (u User) WebAuthnName() string {
	return u.Name
}

func (u User) WebAuthnDisplayName() string {
	return u.DisplayName
}

func (u User) WebAuthnIcon() string {
	return u.Icon
}

func (u User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}
