// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"errors"
)

// ErrStdinIsNotTerminal is returned when Stdin is not an interactive terminal.
var ErrStdinIsNotTerminal = errors.New("stdin is not a terminal")

// ErrConfirmationMismatch is returned when user input does not match the confirmation prompt.
var ErrConfirmationMismatch = errors.New("user input didn't match the confirmation prompt")
