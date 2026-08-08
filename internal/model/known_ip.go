package model

import (
	"database/sql"
	"encoding/json"
	"time"

	"go.yaml.in/yaml/v4"
)

// KnownIP represents a known IP address for a user in the known_ip_addresses table.
type KnownIP struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
	IP       IP     `db:"ip_address"`

	FirstSeen time.Time    `db:"first_seen"`
	LastSeen  time.Time    `db:"last_seen"`
	ExpiresAt sql.NullTime `db:"expires_at"`

	BrowserName    string `db:"browser_name"`
	BrowserVersion string `db:"browser_version"`
	OSName         string `db:"os_name"`
	OSVersion      string `db:"os_version"`
	DeviceType     string `db:"device_type"`
}

// Expires provides ExpiresAt as a *time.Time instead of sql.NullTime. Returns nil if the known IP address never
// expires.
func (k *KnownIP) Expires() *time.Time {
	if k.ExpiresAt.Valid {
		return &k.ExpiresAt.Time
	}

	return nil
}

// MarshalJSON returns the KnownIP in a JSON friendly manner.
func (k *KnownIP) MarshalJSON() (data []byte, err error) {
	return json.Marshal(k.ToData())
}

// MarshalYAML marshals this model into YAML.
func (k *KnownIP) MarshalYAML() (any, error) {
	return k.ToData(), nil
}

// UnmarshalYAML unmarshalls YAML into this model.
func (k *KnownIP) UnmarshalYAML(value *yaml.Node) (err error) {
	o := &KnownIPData{}

	if err = value.Decode(o); err != nil {
		return err
	}

	k.fromData(o)

	return nil
}

// ToData returns the KnownIP as a KnownIPData, suitable for JSON/YAML marshaling.
func (k *KnownIP) ToData() (data KnownIPData) {
	data = KnownIPData{
		Username:       k.Username,
		IP:             k.IP,
		FirstSeen:      k.FirstSeen,
		LastSeen:       k.LastSeen,
		BrowserName:    k.BrowserName,
		BrowserVersion: k.BrowserVersion,
		OSName:         k.OSName,
		OSVersion:      k.OSVersion,
		DeviceType:     k.DeviceType,
	}

	data.ExpiresAt = k.Expires()

	return data
}

func (k *KnownIP) fromData(data *KnownIPData) {
	k.Username = data.Username
	k.IP = data.IP
	k.FirstSeen = data.FirstSeen
	k.LastSeen = data.LastSeen
	k.BrowserName = data.BrowserName
	k.BrowserVersion = data.BrowserVersion
	k.OSName = data.OSName
	k.OSVersion = data.OSVersion
	k.DeviceType = data.DeviceType

	if data.ExpiresAt != nil {
		k.ExpiresAt = sql.NullTime{Valid: true, Time: *data.ExpiresAt}
	}
}

// KnownIPData is used for marshaling/unmarshaling tasks.
type KnownIPData struct {
	Username string `yaml:"username" json:"username" jsonschema:"title=Username" jsonschema_description:"The username this IP address is known for."`
	IP       IP     `yaml:"ip_address" json:"ip_address" jsonschema:"title=IP Address" jsonschema_description:"The known IP address."`

	FirstSeen time.Time  `yaml:"first_seen" json:"first_seen" jsonschema:"title=First Seen" jsonschema_description:"The time this IP address was first seen for this user."`
	LastSeen  time.Time  `yaml:"last_seen" json:"last_seen" jsonschema:"title=Last Seen" jsonschema_description:"The time this IP address was last seen for this user."`
	ExpiresAt *time.Time `yaml:"expires_at,omitempty" json:"expires_at,omitempty" jsonschema:"title=Expires At" jsonschema_description:"The time this known IP address entry expires. Null if it never expires."`

	BrowserName    string `yaml:"browser_name,omitempty" json:"browser_name,omitempty" jsonschema:"title=Browser Name" jsonschema_description:"The name of the browser used with this IP address."`
	BrowserVersion string `yaml:"browser_version,omitempty" json:"browser_version,omitempty" jsonschema:"title=Browser Version" jsonschema_description:"The version of the browser used with this IP address."`
	OSName         string `yaml:"os_name,omitempty" json:"os_name,omitempty" jsonschema:"title=OS Name" jsonschema_description:"The name of the operating system used with this IP address."`
	OSVersion      string `yaml:"os_version,omitempty" json:"os_version,omitempty" jsonschema:"title=OS Version" jsonschema_description:"The version of the operating system used with this IP address."`
	DeviceType     string `yaml:"device_type,omitempty" json:"device_type,omitempty" jsonschema:"title=Device Type" jsonschema_description:"The type of device used with this IP address."`
}

// KnownIPsExport represents a KnownIP export file.
type KnownIPsExport struct {
	KnownIPs []KnownIP `yaml:"known_ips"`
}

// MarshalYAML marshals this model into YAML.
func (export KnownIPsExport) MarshalYAML() (any, error) {
	data := make([]KnownIPData, len(export.KnownIPs))

	for i := range export.KnownIPs {
		data[i] = export.KnownIPs[i].ToData()
	}

	return &KnownIPDataExport{KnownIPs: data}, nil
}

// KnownIPDataExport is the JSON schema friendly representation of a KnownIPsExport.
type KnownIPDataExport struct {
	KnownIPs []KnownIPData `yaml:"known_ips" json:"known_ips" jsonschema:"title=Known IPs" jsonschema_description:"The list of known IP addresses."`
}
