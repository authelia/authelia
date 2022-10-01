package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionDefault(t *testing.T) {
	v := Version()

	assert.Equal(t, "untagged-unknown-dirty (master, unknown)", v)
}

func TestVersion(t *testing.T) {
	var v string

	v = VersionAdv("v4.90.0", "tagged clean", "50d8b4a941c26b89482c94ab324b5a274f9ced66", "master", "")
	assert.Equal(t, "v4.90.0", v)

	v = VersionAdv("v4.90.0", "tagged clean", "50d8b4a941c26b89482c94ab324b5a274f9ced66", "master", "freshports")
	assert.Equal(t, "v4.90.0-freshports", v)

	v = VersionAdv("v4.90.0", "tagged dirty", "50d8b4a941c26b89482c94ab324b5a274f9ced66", "master", "")
	assert.Equal(t, "v4.90.0-dirty", v)

	v = VersionAdv("v4.90.0", "untagged dirty", "50d8b4a941c26b89482c94ab324b5a274f9ced66", "master", "")
	assert.Equal(t, "untagged-v4.90.0-dirty (master, 50d8b4a)", v)

	v = VersionAdv("v4.90.0", "untagged clean", "50d8b4a941c26b89482c94ab324b5a274f9ced66", "master", "")
	assert.Equal(t, "untagged-v4.90.0 (master, 50d8b4a)", v)

	v = VersionAdv("v4.90.0", "untagged clean", "50d8b4a941c26b89482c94ab324b5a274f9ced66", "master", "freshports")
	assert.Equal(t, "untagged-v4.90.0-freshports (master, 50d8b4a)", v)

	v = VersionAdv("v4.90.0", "untagged clean", "", "master", "")
	assert.Equal(t, "untagged-v4.90.0 (master, unknown)", v)

	v = VersionAdv("v4.90.0", "", "50d8b4a941c26b89482c94ab324b5a274f9ced66", "master", "")
	assert.Equal(t, "untagged-v4.90.0-dirty (master, 50d8b4a)", v)
}
