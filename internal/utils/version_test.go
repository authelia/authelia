package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionDefault(t *testing.T) {
	v := Version()

	assert.Equal(t, "untagged-unknown-dirty (master, unknown)", v)
}

func TestVersion(t *testing.T) {
	var v string

	v = version("v4.90.0", "tagged clean", "50d8b4a941c26b89482c94ab324b5a274f9ced66", "master", "")
	assert.Equal(t, "v4.90.0", v)

	v = version("v4.90.0", "tagged clean", "50d8b4a941c26b89482c94ab324b5a274f9ced66", "master", "freshports")
	assert.Equal(t, "v4.90.0-freshports", v)

	v = version("v4.90.0", "tagged dirty", "50d8b4a941c26b89482c94ab324b5a274f9ced66", "master", "")
	assert.Equal(t, "v4.90.0-dirty", v)

	v = version("v4.90.0", "untagged dirty", "50d8b4a941c26b89482c94ab324b5a274f9ced66", "master", "")
	assert.Equal(t, "untagged-v4.90.0-dirty (master, 50d8b4a)", v)

	v = version("v4.90.0", "untagged clean", "50d8b4a941c26b89482c94ab324b5a274f9ced66", "master", "")
	assert.Equal(t, "untagged-v4.90.0 (master, 50d8b4a)", v)

	v = version("v4.90.0", "untagged clean", "50d8b4a941c26b89482c94ab324b5a274f9ced66", "master", "freshports")
	assert.Equal(t, "untagged-v4.90.0-freshports (master, 50d8b4a)", v)

	v = version("v4.90.0", "untagged clean", "", "master", "")
	assert.Equal(t, "untagged-v4.90.0 (master, unknown)", v)

	v = version("v4.90.0", "", "50d8b4a941c26b89482c94ab324b5a274f9ced66", "master", "")
	assert.Equal(t, "untagged-v4.90.0-dirty (master, 50d8b4a)", v)
}

func TestNewSemanticVersion(t *testing.T) {
	var semver SemanticVersion

	semver = NewSemanticVersion("1.2.3")

	assert.Equal(t, "1.2.3", semver.String())
	assert.Equal(t, 1, semver.Major)
	assert.Equal(t, 2, semver.Minor)
	assert.Equal(t, 3, semver.Patch)
	assert.Equal(t, "", semver.PreRelease)
	assert.Len(t, semver.Metadata, 0)

	semver = NewSemanticVersion("v1.2.3-alpha1")

	assert.Equal(t, "1.2.3-alpha1", semver.String())
	assert.Equal(t, 1, semver.Major)
	assert.Equal(t, 2, semver.Minor)
	assert.Equal(t, 3, semver.Patch)
	assert.Equal(t, "alpha1", semver.PreRelease)
	assert.Len(t, semver.Metadata, 0)

	semver = NewSemanticVersion("1.2.3-alpha1+abc.one.two.three")

	assert.Equal(t, "1.2.3-alpha1+abc.one.two.three", semver.String())
	assert.Equal(t, 1, semver.Major)
	assert.Equal(t, 2, semver.Minor)
	assert.Equal(t, 3, semver.Patch)
	assert.Equal(t, "alpha1", semver.PreRelease)

	require.Len(t, semver.Metadata, 4)
	assert.Equal(t, "abc", semver.Metadata[0])
	assert.Equal(t, "one", semver.Metadata[1])
	assert.Equal(t, "two", semver.Metadata[2])
	assert.Equal(t, "three", semver.Metadata[3])
}

func TestSemanticVersionEquals(t *testing.T) {
	assert.True(t, NewSemanticVersion("1.2.3").Equals(NewSemanticVersion("1.2.3")))
	assert.False(t, NewSemanticVersion("1.2.3").Equals(NewSemanticVersion("1.2.0")))
	assert.False(t, NewSemanticVersion("1.2.3").Equals(NewSemanticVersion("1.2.5")))
	assert.False(t, NewSemanticVersion("1.2.3").Equals(NewSemanticVersion("4.2.3")))
	assert.False(t, NewSemanticVersion("1.2.3").Equals(NewSemanticVersion("1.3.3")))
	assert.False(t, NewSemanticVersion("1.2.3").Equals(NewSemanticVersion("1.2.3-alpha1")))
}

func TestSemanticVersionGreater(t *testing.T) {
	assert.True(t, NewSemanticVersion("1.2.4").Greater(NewSemanticVersion("1.2.3")))
	assert.True(t, NewSemanticVersion("1.2.3").Greater(NewSemanticVersion("1.2.3-alpha1")))
	assert.True(t, NewSemanticVersion("2.2.3").Greater(NewSemanticVersion("1.2.3")))
	assert.True(t, NewSemanticVersion("1.3.3").Greater(NewSemanticVersion("1.2.3")))
	assert.True(t, NewSemanticVersion("1.3.3").Greater(NewSemanticVersion("1.2.3-alpha1")))
	assert.True(t, NewSemanticVersion("1.2.3-alpha2").Greater(NewSemanticVersion("1.2.3-alpha1")))
	assert.True(t, NewSemanticVersion("1.2.3-beta1").Greater(NewSemanticVersion("1.2.3-alpha20")))
	assert.False(t, NewSemanticVersion("1.2.3").Greater(NewSemanticVersion("1.2.4")))
	assert.False(t, NewSemanticVersion("1.2.3-alpha1").Greater(NewSemanticVersion("1.2.3")))
	assert.False(t, NewSemanticVersion("1.2.3").Greater(NewSemanticVersion("4.2.3")))
}
