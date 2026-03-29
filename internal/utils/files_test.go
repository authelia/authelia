package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileExists(t *testing.T) {
	dir := t.TempDir()

	filePath := filepath.Join(dir, "testfile")
	require.NoError(t, os.WriteFile(filePath, []byte("data"), 0600))

	testCases := []struct {
		name           string
		path           string
		expectedExists bool
		expectedErr    string
	}{
		{"ShouldReturnTrueForExistingFile", filePath, true, ""},
		{"ShouldReturnErrorForDirectory", dir, false, "path is a directory"},
		{"ShouldReturnFalseForNonExistentFile", filepath.Join(dir, "nonexistent"), false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exists, err := FileExists(tc.path)

			assert.Equal(t, tc.expectedExists, exists)

			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDirectoryExists(t *testing.T) {
	dir := t.TempDir()

	filePath := filepath.Join(dir, "testfile")
	require.NoError(t, os.WriteFile(filePath, []byte("data"), 0600))

	testCases := []struct {
		name           string
		path           string
		expectedExists bool
		expectedErr    string
	}{
		{"ShouldReturnTrueForExistingDirectory", dir, true, ""},
		{"ShouldReturnErrorForFile", filePath, false, "path is a file"},
		{"ShouldReturnFalseForNonExistentDirectory", filepath.Join(dir, "nonexistent"), false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exists, err := DirectoryExists(tc.path)

			assert.Equal(t, tc.expectedExists, exists)

			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPathExists(t *testing.T) {
	dir := t.TempDir()

	filePath := filepath.Join(dir, "testfile")
	require.NoError(t, os.WriteFile(filePath, []byte("data"), 0600))

	testCases := []struct {
		name           string
		path           string
		expectedExists bool
	}{
		{"ShouldReturnTrueForExistingFile", filePath, true},
		{"ShouldReturnTrueForExistingDirectory", dir, true},
		{"ShouldReturnFalseForNonExistentFile", filepath.Join(dir, "nonexistent"), false},
		{"ShouldReturnFalseForNonExistentDirectory", filepath.Join(dir, "nonexistent", "dir"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exists, err := PathExists(tc.path)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedExists, exists)
		})
	}
}
