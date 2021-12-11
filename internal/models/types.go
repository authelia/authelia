package models

import (
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
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

// IP is a type specific for storage of a net.IP in the database which can't be NULL.
type IP struct {
	IP net.IP
}

// Value is the IP implementation of the databases/sql driver.Valuer.
func (ip IP) Value() (value driver.Value, err error) {
	if ip.IP == nil {
		return nil, errors.New("cannot value nil IP to driver.Value")
	}

	return driver.Value(ip.IP.String()), nil
}

// Scan is the IP implementation of the sql.Scanner.
func (ip *IP) Scan(src interface{}) (err error) {
	if src == nil {
		return errors.New("cannot scan nil to type IP")
	}

	var value string

	switch v := src.(type) {
	case string:
		value = v
	case []byte:
		value = string(v)
	default:
		return fmt.Errorf("invalid type %T for IP %v", src, src)
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
		return driver.Value(nil), nil
	}

	return driver.Value(ip.IP.String()), nil
}

// Scan is the NullIP implementation of the sql.Scanner.
func (ip *NullIP) Scan(src interface{}) (err error) {
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
		return fmt.Errorf("invalid type %T for NullIP %v", src, src)
	}

	ip.IP = net.ParseIP(value)

	return nil
}

// NewHexadecimal returns a new  Hexadecimal given some bytes.
func NewHexadecimal(value []byte) (hexadecimal Hexadecimal) {
	return Hexadecimal{value: value}
}

// Hexadecimal represents bytes stored in the database as a hexadecimal string.
type Hexadecimal struct {
	value []byte
}

// String returns the encoded string value of the bytes.
func (h Hexadecimal) String() string {
	return hex.EncodeToString(h.value)
}

// Bytes returns the value.
func (h Hexadecimal) Bytes() []byte {
	return h.value
}

// Value implements the driver.Valuer.
func (h Hexadecimal) Value() (value driver.Value, err error) {
	return hex.EncodeToString(h.value), nil
}

// Scan implements the sql.Scanner.
func (h *Hexadecimal) Scan(src interface{}) (err error) {
	if src == nil {
		return errors.New("cannot scan nil to type Hexadecimal")
	}

	switch v := src.(type) {
	case string:
		if h.value, err = hex.DecodeString(v); err != nil {
			return fmt.Errorf("can't convert string to hex: %w", err)
		}
	case []byte:
		if h.value, err = hex.DecodeString(string(v)); err != nil {
			return fmt.Errorf("can't convert byte[] to hex: %w", err)
		}
	default:
		return fmt.Errorf("invalid type %T for Hexadecimal %v", src, src)
	}

	return nil
}

// StartupCheck represents a provider that has a startup check.
type StartupCheck interface {
	StartupCheck() (err error)
}
