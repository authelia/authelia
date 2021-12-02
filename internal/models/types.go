package models

import (
	"database/sql/driver"
	"fmt"
	"net"
)

// NewIPAddressFromString converts a string into an IPAddress.
func NewIPAddressFromString(ip string) (ipAddress IPAddress) {
	actualIP := net.ParseIP(ip)
	return IPAddress{IP: &actualIP}
}

// NewIPAddress creates an IPAddress from net.IP.
func NewIPAddress(ip net.IP) (ipAddress IPAddress) {
	return IPAddress{IP: &ip}
}

// IPAddress is a type specific for storage of a net.IP in the database.
type IPAddress struct {
	*net.IP
}

// Value is the IPAddress implementation of the databases/sql driver.Valuer.
func (ip IPAddress) Value() (value driver.Value, err error) {
	if ip.IP == nil {
		return driver.Value(nil), nil
	}

	return driver.Value(ip.IP.String()), nil
}

// Scan is the IPAddress implementation of the sql.Scanner.
func (ip *IPAddress) Scan(src interface{}) (err error) {
	if src == nil {
		ip.IP = nil
		return nil
	}

	var value string

	switch v := src.(type) {
	case string:
		value = v
	default:
		return fmt.Errorf("invalid type %T for IPAddress %v", src, src)
	}

	*ip.IP = net.ParseIP(value)

	return nil
}

// StartupCheck represents a provider that has a startup check.
type StartupCheck interface {
	StartupCheck() (err error)
}
