package authentication

import (
	"time"

	"github.com/authelia/authelia/v4/internal/model"
)

// CachedUserDetails represent the cached details retrieved for a given user.
type CachedUserDetails struct {
	updated time.Time
	details *model.UserDetails
}
