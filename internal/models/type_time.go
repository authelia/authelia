package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Time allows sqlx's StructScan to Scan the time from int64 and to int64.
type Time struct {
	time.Time
}

// Value returns the value for the database/sql driver.
func (t Time) Value() (value driver.Value, err error) {
	return driver.Value(t.Time.Unix()), nil
}

// Scan allows the database/sql driver to scan the int64 into a time.Time.
func (t *Time) Scan(src interface{}) (err error) {
	var value int64

	switch s := src.(type) {
	case int64:
		value = s
	case nil:
		value = 0
	default:
		return fmt.Errorf("invalid type %T for Time", src)
	}

	*t = Time{
		time.Unix(value, 0),
	}

	return nil
}
