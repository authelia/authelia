package utils

import (
	"fmt"
	"strings"
)

// BuildTag is replaced by LDFLAGS at build time with the latest tag at or before the current commit.
var BuildTag = ""

// BuildStateTag is replaced by LDFLAGS at build time with `tagged` or `untagged` depending on if the commit is tagged.
var BuildStateTag = "tagged"

// BuildStateExtra is replaced by LDFLAGS at build time with a blank string or `dirty` if the working tree is dirty.
var BuildStateExtra = ""

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

// Version returns the short version.
//
// The format of the string is `<latest tag>-<short commit>-<extra>`. Where short commit and the hyphen are only present
// when the commit is not tagged, and the extra is only present if the extra is not blank. Extra is usually used to
// communicate if the working tree is dirty. Though it can be used by ports to include the port name.
//
func Version() (version string) {
	if BuildStateTag == tagged {
		return BuildTag
	}

	b := strings.Builder{}

	b.WriteString("untagged-")
	b.WriteString(BuildTag)

	if BuildStateExtra != "" {
		b.WriteRune('-')
		b.WriteString(BuildStateExtra)
	}

	b.WriteString(fmt.Sprintf(" (%s, %s)", BuildBranch, CommitShort()))

	return b.String()
}
