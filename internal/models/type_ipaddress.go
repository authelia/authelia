package models

import (
	"database/sql/driver"
	"errors"
	"net"
)

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

	switch src.(type) {
	case string:
		value = src.(string)
	default:
		return errors.New("invalid type for IPAddress")
	}

	*ip.IP = net.ParseIP(value)

	return nil
}
