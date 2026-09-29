// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package utils

import (
	"os/exec"
	"syscall"
)

func setProcessGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	cmd.SysProcAttr.Setpgid = true
}

func signalProcessGroup(cmd *exec.Cmd, sig syscall.Signal) error {
	return syscall.Kill(-cmd.Process.Pid, sig)
}
