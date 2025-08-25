package model

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"image"
	"net/url"
	"strconv"
	"time"

	"github.com/authelia/otp"
	"go.yaml.in/yaml/v4"
)

type TOTPOptions struct {
	Algorithm  string   `json:"algorithm"`
	Algorithms []string `json:"algorithms"`

	Length  int   `json:"length"`
	Lengths []int `json:"lengths"`

	Period  int   `json:"period"`
	Periods []int `json:"periods"`
}

// TOTPConfiguration represents a users TOTP configuration row in the database.
type TOTPConfiguration struct {
	ID         int          `db:"id"`
	CreatedAt  time.Time    `db:"created_at"`
	LastUsedAt sql.NullTime `db:"last_used_at"`
	Username   string       `db:"username"`
	Issuer     string       `db:"issuer"`
	Algorithm  string       `db:"algorithm"`
	Digits     uint32       `db:"digits"`
	Period     uint         `db:"period"`
	Secret     []byte       `db:"secret"`
}

// TOTPConfigurationJSON is the JSON representation for a TOTPConfiguration.
type TOTPConfigurationJSON struct {
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	Issuer     string     `json:"issuer"`
	Algorithm  string     `json:"algorithm"`
	Digits     uint32     `json:"digits"`
	Period     uint       `json:"period"`
}

// MarshalJSON returns the TOTPConfiguration in a JSON friendly manner.
func (c TOTPConfiguration) MarshalJSON() (data []byte, err error) {
	o := TOTPConfigurationJSON{
		CreatedAt: c.CreatedAt,
		Issuer:    c.Issuer,
		Algorithm: c.Algorithm,
		Digits:    c.Digits,
		Period:    c.Period,
	}

	if c.LastUsedAt.Valid {
		o.LastUsedAt = &c.LastUsedAt.Time
	}

	return json.Marshal(o)
}

// HistorySince provides a reasonably accurate window for previously successful attempts to check for history.
func (c *TOTPConfiguration) HistorySince(now time.Time, skew *int) time.Time {
	var s int

	if skew == nil {
		s = 2
	} else {
		s = *skew + 2
	}

	// TODO: Adjust the logic here to not require the lint comment.
	//nolint:gosec // Safe as the values set are always convertable to int64.
	return now.Add(-time.Second * time.Duration(c.Period) * time.Duration(s))
}

// LastUsed provides LastUsedAt as a *time.Time instead of sql.NullTime.
func (c *TOTPConfiguration) LastUsed() *time.Time {
	if c.LastUsedAt.Valid {
		value := time.Unix(c.LastUsedAt.Time.Unix(), int64(c.LastUsedAt.Time.Nanosecond()))

		return &value
	}

	return nil
}

// URI shows the configuration in the URI representation.
func (c *TOTPConfiguration) URI() (uri string) {
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

// UpdateSignInInfo adjusts the values of the TOTPConfiguration after a sign in.
func (c *TOTPConfiguration) UpdateSignInInfo(now time.Time) {
	c.LastUsedAt = sql.NullTime{Time: now, Valid: true}
}

// Key returns the *otp.Key using TOTPConfiguration.URI with otp.NewKeyFromURL.
func (c *TOTPConfiguration) Key() (key *otp.Key, err error) {
	return otp.NewKeyFromURL(c.URI())
}

// Image returns the image.Image of the TOTPConfiguration using the Image func from the return of TOTPConfiguration.Key.
func (c *TOTPConfiguration) Image(width, height int) (img image.Image, err error) {
	var key *otp.Key

	if key, err = c.Key(); err != nil {
		return nil, err
	}

	return key.Image(width, height)
}

// ToData converts this TOTPConfiguration into the data format for exporting etc.
func (c *TOTPConfiguration) ToData() TOTPConfigurationData {
	return TOTPConfigurationData{
		CreatedAt:  c.CreatedAt,
		LastUsedAt: c.LastUsed(),
		Username:   c.Username,
		Issuer:     c.Issuer,
		Algorithm:  c.Algorithm,
		Digits:     c.Digits,
		Period:     c.Period,
		Secret:     base64.StdEncoding.EncodeToString(c.Secret),
	}
}

// MarshalYAML marshals this model into YAML.
func (c *TOTPConfiguration) MarshalYAML() (any, error) {
	return c.ToData(), nil
}

// UnmarshalYAML unmarshalls YAML into this model.
func (c *TOTPConfiguration) UnmarshalYAML(value *yaml.Node) (err error) {
	o := &TOTPConfigurationData{}

	if err = value.Decode(o); err != nil {
		return err
	}

	if c.Secret, err = base64.StdEncoding.DecodeString(o.Secret); err != nil {
		return err
	}

	c.CreatedAt = o.CreatedAt
	c.Username = o.Username
	c.Issuer = o.Issuer
	c.Algorithm = o.Algorithm
	c.Digits = o.Digits
	c.Period = o.Period

	if o.LastUsedAt != nil {
		c.LastUsedAt = sql.NullTime{Valid: true, Time: *o.LastUsedAt}
	}

	return nil
}

// TOTPConfigurationData is used for marshalling/unmarshalling tasks.
type TOTPConfigurationData struct {
	CreatedAt  time.Time  `yaml:"created_at" json:"created_at" jsonschema:"title=Created At" jsonschema_description:"The time the configuration was created."`
	LastUsedAt *time.Time `yaml:"last_used_at" json:"last_used_at" jsonschema:"title=Last Used At" jsonschema_description:"The time the configuration was last used at."`
	Username   string     `yaml:"username" json:"username" jsonschema:"title=Username" jsonschema_description:"The username of the user this configuration belongs to."`
	Issuer     string     `yaml:"issuer" json:"issuer" jsonschema:"title=Issuer" jsonschema_description:"The issuer name this was generated with."`
	Algorithm  string     `yaml:"algorithm" json:"algorithm" jsonschema:"title=Algorithm" jsonschema_description:"The algorithm this configuration uses."`
	Digits     uint32     `yaml:"digits" json:"digits" jsonschema:"title=Digits" jsonschema_description:"The number of digits this configuration uses."`
	Period     uint       `yaml:"period" json:"period" jsonschema:"title=Period" jsonschema_description:"The period of time this configuration uses."`
	Secret     string     `yaml:"secret" json:"secret" jsonschema:"title=Secret" jsonschema_description:"The secret shared key for this configuration."`
}

// TOTPConfigurationDataExport represents a TOTPConfiguration export file.
type TOTPConfigurationDataExport struct {
	TOTPConfigurations []TOTPConfigurationData `yaml:"totp_configurations" json:"totp_configurations" jsonschema:"title=TOTP Configurations" jsonschema_description:"The list of TOTP configurations."`
}

// TOTPConfigurationExport represents a TOTPConfiguration export file.
type TOTPConfigurationExport struct {
	TOTPConfigurations []TOTPConfiguration `yaml:"totp_configurations"`
}

// ToData converts this TOTPConfigurationExport into a TOTPConfigurationDataExport.
func (export TOTPConfigurationExport) ToData() TOTPConfigurationDataExport {
	data := TOTPConfigurationDataExport{
		TOTPConfigurations: make([]TOTPConfigurationData, len(export.TOTPConfigurations)),
	}

	for i, config := range export.TOTPConfigurations {
		data.TOTPConfigurations[i] = config.ToData()
	}

	return data
}

// MarshalYAML marshals this model into YAML.
func (export TOTPConfigurationExport) MarshalYAML() (any, error) {
	return export.ToData(), nil
}
