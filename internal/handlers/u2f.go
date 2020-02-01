package handlers

import (
	"crypto/elliptic"

	"github.com/tstranex/u2f"
)

type U2FVerifier interface {
	Verify(keyHandle []byte, publicKey []byte, signResponse u2f.SignResponse, challenge u2f.Challenge) error
}

type U2FVerifierImpl struct{}

func (uv *U2FVerifierImpl) Verify(keyHandle []byte, publicKey []byte,
	signResponse u2f.SignResponse, challenge u2f.Challenge) error {
	var registration u2f.Registration
	registration.KeyHandle = keyHandle
	x, y := elliptic.Unmarshal(elliptic.P256(), publicKey)
	registration.PubKey.Curve = elliptic.P256()
	registration.PubKey.X = x
	registration.PubKey.Y = y

	// TODO(c.michaud): store the counter to help detecting cloned U2F keys.
	_, err := registration.Authenticate(
		signResponse, challenge, 0)
	return err
}
