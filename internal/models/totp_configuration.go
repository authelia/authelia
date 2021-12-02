package models

import (
	"net/url"
	"strconv"
)

// TOTPConfiguration represents a users TOTP configuration row in the database.
type TOTPConfiguration struct {
	ID        int    `db:"id" json:"-"`
	Username  string `db:"username" json:"-"`
	Issuer    string `db:"issuer" json:"-"`
	Algorithm string `db:"algorithm" json:"-"`
	Digits    uint   `db:"digits" json:"digits"`
	Period    uint   `db:"period" json:"period"`
	Secret    []byte `db:"secret" json:"-"`
}

// URI shows the configuration in the URI representation.
func (c TOTPConfiguration) URI() (uri string) {
	v := url.Values{}
	v.Set("secret", string(c.Secret))
	v.Set("issuer", c.Issuer)
	v.Set("period", strconv.FormatUint(uint64(c.Period), 10))
	v.Set("algorithm", c.Algorithm)
	v.Set("digits", strconv.Itoa(int(c.Digits)))

	u := url.URL{
		Scheme:   "otpauth",
		Host:     "totp",
		Path:     "/" + c.Issuer + ":" + c.Username,
		RawQuery: v.Encode(),
	}

	return u.String()
}
