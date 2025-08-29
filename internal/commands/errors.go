package commands

import (
	"errors"
)

// ErrStdinIsNotTerminal is returned when Stdin is not an interactive terminal.
var ErrStdinIsNotTerminal = errors.New("stdin is not a terminal")

// ErrConfirmationMismatch is returned when user input does not match the confirmation prompt.
var ErrConfirmationMismatch = errors.New("user input didn't match the confirmation prompt")

// ErrConfigCreated is returned when the configuration file is created.
var ErrConfigCreated = errors.New("configuration created")
