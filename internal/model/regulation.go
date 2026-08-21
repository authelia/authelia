package model

import (
	"database/sql"
	"time"
)

// AuthenticationAttempt represents an authentication attempt row in the database.
type AuthenticationAttempt struct {
	ID            int       `db:"id"`
	Time          time.Time `db:"time"`
	Successful    bool      `db:"successful"`
	Banned        bool      `db:"banned"`
	Username      string    `db:"username"`
	Type          string    `db:"auth_type"`
	RemoteIP      NullIP    `db:"remote_ip"`
	RequestURI    string    `db:"request_uri"`
	RequestMethod string    `db:"request_method"`
}

// RegulationRecord represents an authentication attempt record used by the regulator.
type RegulationRecord struct {
	Time       time.Time `db:"time"`
	Successful bool      `db:"successful"`
}

// BannedUser represents a banned user row in the storage provider.
type BannedUser struct {
	ID       int            `db:"id"`
	Time     time.Time      `db:"time"`
	Expires  sql.NullTime   `db:"expires"`
	Expired  sql.NullTime   `db:"expired"`
	Revoked  bool           `db:"revoked"`
	Username string         `db:"username"`
	Source   string         `db:"source"`
	Reason   sql.NullString `db:"reason"`
}

// BannedIP represents a banned IP row in the storage provider.
type BannedIP struct {
	ID      int            `db:"id"`
	Time    time.Time      `db:"time"`
	Expires sql.NullTime   `db:"expires"`
	Expired sql.NullTime   `db:"expired"`
	Revoked bool           `db:"revoked"`
	IP      IP             `db:"ip"`
	Source  string         `db:"source"`
	Reason  sql.NullString `db:"reason"`
}
