package models

import (
	"database/sql/driver"
	"fmt"
	"net"
)

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
