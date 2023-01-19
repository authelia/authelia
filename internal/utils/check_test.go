package utils

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheckUntilPredicateOk(t *testing.T) {
	interval := time.Second * 1
	timeout := time.Second * 5

	err := CheckUntil(interval, timeout, func() (bool, error) {
		return true, nil
	})
	assert.NoError(t, err, "")
}

func TestCheckUntilPredicateError(t *testing.T) {
	interval := time.Second * 1
	timeout := time.Second * 5

	theError := errors.New("some error")

	err := CheckUntil(interval, timeout, func() (bool, error) {
		return false, theError
	})
	assert.ErrorIs(t, err, theError)
}

func TestCheckUntilPredicateTimeout(t *testing.T) {
	interval := 1 * time.Second
	timeout := 3 * time.Second

	err := CheckUntil(interval, timeout, func() (bool, error) {
		return false, nil
	})
	assert.ErrorContains(t, err, "timeout of 3s reached")
}
