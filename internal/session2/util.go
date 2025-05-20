package session2

import (
	"github.com/authelia/authelia/v4/internal/random"
)

func genSessionID(provider random.Provider) []byte {
	return provider.BytesCustom(32, []byte(random.CharSetAlphaNumeric))
}
