package commands

import (
	"errors"
)

// ErrStdinIsNotTerminal is returned when Stdin is not an interactive terminal.
var ErrStdinIsNotTerminal = errors.New("stdin is not a terminal")
