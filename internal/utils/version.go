package utils

import (
	"fmt"
	"strings"
)

// BuildTag is replaced by LDFLAGS at build time with the latest tag at or before the current commit.
var BuildTag = ""

// BuildState is replaced by LDFLAGS at build time with `tagged` or `untagged` depending on if the commit is tagged, and
// `clean` or `dirty` depending on the working tree state. For example if the commit was tagged and the working tree
// was dirty it would be "tagged dirty". This is used to determine the version string output mode.
var BuildState = "untagged dirty"

// BuildExtra is replaced by LDFLAGS at build time with a blank string by default. People porting Authelia can use this
// to add a suffix to their versions.
var BuildExtra = ""

// BuildDate is replaced by LDFLAGS at build time with the date the build started.
var BuildDate = ""

// BuildCommit is replaced by LDFLAGS at build time with the current commit.
var BuildCommit = "unknown"

// BuildBranch is replaced by LDFLAGS at build time with the current branch.
var BuildBranch = "master"

// BuildNumber is replaced by LDFLAGS at build time with the CI build number.
var BuildNumber = "0"

// BuildArch is replaced by LDFLAGS at build time with the CI build arch.
var BuildArch = ""

// CommitShort loops through the BuildCommit chars and safely writes the first 7 to a string builder and returns it.
func CommitShort() (commit string) {
	if BuildCommit == "" {
		return unknown
	}

	b := strings.Builder{}

	for i, r := range BuildCommit {
		b.WriteRune(r)

		if i >= 6 {
			break
		}
	}

	return b.String()
}

// Version returns the Authelia version.
//
// The format of the string is dependent on the values in BuildState. If untagged or dirty are not present it returns
// the BuildTag i.e. v1.0.0. If this is not true the following is the format:
// untagged-<BuildTag>-dirty-<BuildExtra> (<BuildBranch>, <BuildCommit>).
//
func Version() (version string) {
	b := strings.Builder{}

	states := strings.Split(BuildState, " ")

	if IsStringInSlice(clean, states) && IsStringInSlice(tagged, states) {
		b.WriteString(BuildTag)

		if BuildExtra != "" {
			b.WriteRune('-')
			b.WriteString(BuildExtra)
		}

		return b.String()
	}

	if IsStringInSlice(untagged, states) {
		b.WriteString("untagged-")
	}

	b.WriteString(BuildTag)

	if IsStringInSlice(dirty, states) {
		b.WriteString("-dirty")
	}

	if BuildExtra != "" {
		b.WriteRune('-')
		b.WriteString(BuildExtra)
	}

	b.WriteString(fmt.Sprintf(" (%s, %s)", BuildBranch, CommitShort()))

	return b.String()
}
