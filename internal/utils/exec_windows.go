// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package utils

import (
	"os/exec"
	"syscall"
)

func setProcessGroup(_ *exec.Cmd) {}

func signalProcessGroup(cmd *exec.Cmd, sig syscall.Signal) error {
	if sig != syscall.SIGKILL {
		return nil
	}

	return cmd.Process.Kill()
}
