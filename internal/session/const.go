package session

import (
	"time"
)

const (
	randomSessionChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_!#$%^*"

	hkdfKeyInfoCodec = "authelia:kdf:session:codec:encryption_key:v1"
)

const (
	// cookieDeletionOffset is how far in the past a deletion cookie expires, which must be sufficient to account for
	// clock skew between this server and the user agent.
	cookieDeletionOffset = time.Hour * 24
)

var (
	expireUnlimited time.Time
)
