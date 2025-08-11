package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldCheckIfFileExists(t *testing.T) {
	exists, err := FileExists("../../README.md")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = FileExists("../../")
	assert.EqualError(t, err, "path is a directory")
	assert.False(t, exists)

	exists, err = FileExists("../../NOTAFILE.md")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestShouldCheckIfDirectoryExists(t *testing.T) {
	exists, err := DirectoryExists("../../")

	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = DirectoryExists("../../README.md")
	assert.EqualError(t, err, "path is a file")
	assert.False(t, exists)

	exists, err = DirectoryExists("../../NOTADIRECTORY/")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestShouldCheckIfPathExists(t *testing.T) {
	exists, err := PathExists("../../README.md")

	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = PathExists("../../")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = PathExists("../../NOTAFILE.md")
	assert.NoError(t, err)
	assert.False(t, exists)

	exists, err = PathExists("../../NOTADIRECTORY/")
	assert.NoError(t, err)
	assert.False(t, exists)
}
