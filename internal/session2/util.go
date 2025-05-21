package session2

import (
	"github.com/authelia/authelia/v4/internal/random"
	"time"
)

func genSessionID(provider random.Provider) []byte {
	return provider.BytesCustom(32, []byte(random.CharSetAlphaNumeric))
}

func storeExp(expiration time.Duration) time.Duration {
	if expiration == -1 {
		return 2 * 24 * time.Hour
	}

	return expiration
}
