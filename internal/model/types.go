package model

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"fmt"
	"net"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewIP easily constructs a new IP.
func NewIP(value net.IP) (ip IP) {
	return IP{IP: value}
}

// NewNullIP easily constructs a new NullIP.
func NewNullIP(value net.IP) (ip NullIP) {
	return NullIP{IP: value}
}

// NewNullIPFromString easily constructs a new NullIP from a string.
func NewNullIPFromString(value string) (ip NullIP) {
	if value == "" {
		return ip
	}

	return NullIP{IP: net.ParseIP(value)}
}

// NewBase64 returns a new Base64.
func NewBase64(data []byte) Base64 {
	return Base64{data: data}
}

// IP is a type specific for storage of a net.IP in the database which can't be NULL.
type IP struct {
	IP net.IP
}

// Value is the IP implementation of the databases/sql driver.Valuer.
func (ip IP) Value() (value driver.Value, err error) {
	if ip.IP == nil {
		return nil, fmt.Errorf(errFmtValueNil, ip)
	}

	return ip.IP.String(), nil
}

// Scan is the IP implementation of the sql.Scanner.
func (ip *IP) Scan(src any) (err error) {
	if src == nil {
		return fmt.Errorf(errFmtScanNil, ip)
	}

	var value string

	switch v := src.(type) {
	case string:
		value = v
	case []byte:
		value = string(v)
	default:
		return fmt.Errorf(errFmtScanInvalidType, ip, src, src)
	}

	ip.IP = net.ParseIP(value)

	return nil
}

// NullIP is a type specific for storage of a net.IP in the database which can also be NULL.
type NullIP struct {
	IP net.IP
}

// Value is the NullIP implementation of the databases/sql driver.Valuer.
func (ip NullIP) Value() (value driver.Value, err error) {
	if ip.IP == nil {
		return nil, nil
	}

	return ip.IP.String(), nil
}

// Scan is the NullIP implementation of the sql.Scanner.
func (ip *NullIP) Scan(src any) (err error) {
	if src == nil {
		ip.IP = nil
		return nil
	}

	var value string

	switch v := src.(type) {
	case string:
		value = v
	case []byte:
		value = string(v)
	default:
		return fmt.Errorf(errFmtScanInvalidType, ip, src, src)
	}

	ip.IP = net.ParseIP(value)

	return nil
}

// Base64 saves bytes to the database as a base64 encoded string.
type Base64 struct {
	data []byte
}

// String returns the Base64 string encoded as base64.
func (b Base64) String() string {
	return base64.StdEncoding.EncodeToString(b.data)
}

// Bytes returns the Base64 string encoded as bytes.
func (b Base64) Bytes() []byte {
	return b.data
}

// Value is the Base64 implementation of the databases/sql driver.Valuer.
func (b Base64) Value() (value driver.Value, err error) {
	return b.String(), nil
}

// Scan is the Base64 implementation of the sql.Scanner.
func (b *Base64) Scan(src any) (err error) {
	if src == nil {
		return fmt.Errorf(errFmtScanNil, b)
	}

	switch v := src.(type) {
	case string:
		if b.data, err = base64.StdEncoding.DecodeString(v); err != nil {
			return fmt.Errorf(errFmtScanInvalidTypeErr, b, src, src, err)
		}
	case []byte:
		if b.data, err = base64.StdEncoding.DecodeString(string(v)); err != nil {
			b.data = v
		}
	default:
		return fmt.Errorf(errFmtScanInvalidType, b, src, src)
	}

	return nil
}

// StartupCheck represents a provider that has a startup check.
type StartupCheck interface {
	StartupCheck() (err error)
}

// StringSlicePipeDelimited is a string slice that is stored in the database delimited by pipes.
type StringSlicePipeDelimited []string

// Scan is the StringSlicePipeDelimited implementation of the sql.Scanner.
func (s *StringSlicePipeDelimited) Scan(value any) (err error) {
	var nullStr sql.NullString

	if err = nullStr.Scan(value); err != nil {
		return err
	}

	if nullStr.Valid {
		*s = utils.StringSplitDelimitedEscaped(nullStr.String, '|')
	}

	return nil
}

// Value is the StringSlicePipeDelimited implementation of the databases/sql driver.Valuer.
func (s StringSlicePipeDelimited) Value() (driver.Value, error) {
	return utils.StringJoinDelimitedEscaped(s, '|'), nil
}

// Context is a commonly used context.Context within Authelia.
type Context interface {
	context.Context

	GetClock() clock.Provider
	RemoteIP() net.IP
	GetRandom() random.Provider
}
