// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package utils

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldKillDescendantsOnTimeout(t *testing.T) {
	pidfile := filepath.Join(t.TempDir(), "pid")

	cmd := Shell("sleep 30 & echo $! > " + pidfile + "; wait")

	err := RunCommandWithTimeout(cmd, time.Second)
	require.ErrorIs(t, err, ErrTimeoutReached)

	data, err := os.ReadFile(pidfile)
	require.NoError(t, err)

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		return errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) || isZombie(pid)
	}, 5*time.Second, 50*time.Millisecond, "the command's child %d outlived the timeout", pid)
}

func isZombie(pid int) bool {
	stat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return false
	}

	fields := strings.Fields(string(stat[bytes.LastIndexByte(stat, ')')+1:]))

	return len(fields) > 0 && fields[0] == "Z"
}
