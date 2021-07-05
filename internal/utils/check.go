package utils

import (
	"fmt"
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
			return fmt.Errorf("timeout of %ds reached", int64(timeout/time.Second))
		}
	}
}
