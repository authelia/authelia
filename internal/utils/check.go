package utils

import (
	"fmt"
	"reflect"
	"time"
)

// CheckUntil regularly check a predicate until it's true or time out is reached.
func CheckUntil(interval time.Duration, timeout time.Duration, predicate func() (bool, error)) error {
	for {
		select {
		case <-time.After(interval):
			predTrue, err := predicate()
			if predTrue {
				return nil
			}

			if err != nil {
				return err
			}
		case <-time.After(timeout):
			return fmt.Errorf("Timeout of %ds reached", int64(timeout/time.Second))
		}
	}
}

// IsNil checks if an interface is nil.
func IsNil(object interface{}) bool {
	if object == nil {
		return true
	}

	switch reflect.TypeOf(object).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(object).IsNil()
	}

	return false
}

// CountNil counts the number of nil objects.
func CountNil(objects ...interface{}) (count int) {
	for _, object := range objects {
		if !IsNil(object) {
			count++
		}
	}

	return count
}
