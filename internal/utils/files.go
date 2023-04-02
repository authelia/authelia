// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"errors"
	"os"
)

// FileExists returns true if the given path exists and is a file.
func FileExists(path string) (exists bool, err error) {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return false, errors.New("path is a directory")
		}

		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// DirectoryExists returns true if the given path exists and is a directory.
func DirectoryExists(path string) (exists bool, err error) {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return true, nil
		}

		return false, errors.New("path is a file")
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// PathExists returns true if the given path exists.
func PathExists(path string) (exists bool, err error) {
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}
