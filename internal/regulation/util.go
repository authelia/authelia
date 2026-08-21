package regulation

import (
	"database/sql"
	"time"
)

func returnBanResult(b BanType, v string, t sql.NullTime) (ban BanType, value string, expires *time.Time, err error) {
	if t.Valid {
		expires = &t.Time
	}

	return b, v, expires, ErrUserIsBanned
}

// FormatExpiresLong returns the long form representation of the given expiration time.
func FormatExpiresLong(expires *time.Time) string {
	if expires == nil {
		return "never expires"
	}

	return expires.Format("3:04:05PM on January 2 2006 (-07:00)")
}

// FormatExpiresShort returns the short form representation of the given expiration time.
func FormatExpiresShort(expires sql.NullTime) string {
	if !expires.Valid {
		return "never"
	}

	return expires.Time.Format(time.DateTime)
}
