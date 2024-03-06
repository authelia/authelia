package model

import (
	"database/sql"
	"net"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// NewIdentityVerification creates a new IdentityVerification from a given username and action.
func NewIdentityVerification(jti uuid.UUID, username, action string, ip net.IP, expiration time.Duration) (verification IdentityVerification) {
	return IdentityVerification{
		JTI:       jti,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(expiration),
		Action:    action,
		Username:  username,
		IssuedIP:  NewIP(ip),
	}
}

// IdentityVerification represents an identity verification row in the database.
type IdentityVerification struct {
	ID         int          `db:"id"`
	JTI        uuid.UUID    `db:"jti"`
	IssuedAt   time.Time    `db:"iat"`
	IssuedIP   IP           `db:"issued_ip"`
	ExpiresAt  time.Time    `db:"exp"`
	Action     string       `db:"action"`
	Username   string       `db:"username"`
	ConsumedAt sql.NullTime `db:"consumed"`
	ConsumedIP NullIP       `db:"consumed_ip"`
	RevokedAt  sql.NullTime `db:"revoked"`
	RevokedIP  NullIP       `db:"revoked_ip"`
}

// ToIdentityVerificationClaim converts the IdentityVerification into a IdentityVerificationClaim.
func (v IdentityVerification) ToIdentityVerificationClaim() (claim *IdentityVerificationClaim) {
	return &IdentityVerificationClaim{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        v.JTI.String(),
			Issuer:    "Authelia",
			IssuedAt:  jwt.NewNumericDate(v.IssuedAt),
			ExpiresAt: jwt.NewNumericDate(v.ExpiresAt),
		},
		Action:   v.Action,
		Username: v.Username,
	}
}

// IdentityVerificationClaim custom claim for specifying the action claim.
// The action can be to register a TOTP device, a U2F device or reset one's password.
type IdentityVerificationClaim struct {
	jwt.RegisteredClaims

	// The action this token has been crafted for.
	Action string `json:"action"`
	// The user this token has been crafted for.
	Username string `json:"username"`
}

// ToIdentityVerification converts the IdentityVerificationClaim into a IdentityVerification.
func (v IdentityVerificationClaim) ToIdentityVerification() (verification *IdentityVerification, err error) {
	jti, err := uuid.Parse(v.ID)
	if err != nil {
		return nil, err
	}

	return &IdentityVerification{
		JTI:       jti,
		Username:  v.Username,
		Action:    v.Action,
		ExpiresAt: v.ExpiresAt.Time,
	}, nil
}
