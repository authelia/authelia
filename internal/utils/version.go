package utils

import (
	"fmt"
	"strings"
)

// BuildTag is replaced by LDFLAGS at build time with the latest tag at or before the current commit.
var BuildTag = "unknown"

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

// Version returns the Authelia version.
//
// The format of the string is dependent on the values in BuildState. If tagged and clean are present it returns the
// BuildTag i.e. v1.0.0. If dirty and tagged are present it returns <BuildTag>-dirty. Otherwise, the following is the
// format: untagged-<BuildTag>-dirty-<BuildExtra> (<BuildBranch>, <BuildCommit>).
func Version() (versionString string) {
	return VersionAdv(BuildTag, BuildState, BuildCommit, BuildBranch, BuildExtra)
}

// VersionAdv takes inputs to generate the version.
func VersionAdv(tag, state, commit, branch, extra string) (version string) {
	b := strings.Builder{}

	states := strings.Split(state, " ")

	isClean := IsStringInSlice(clean, states)
	isTagged := IsStringInSlice(tagged, states)

	if isClean && isTagged {
		b.WriteString(tag)

		if extra != "" {
			b.WriteRune('-')
			b.WriteString(extra)
		}

		return b.String()
	}

	if isTagged && !isClean {
		b.WriteString(tag)
		b.WriteString("-dirty")

		return b.String()
	}

	if !isTagged {
		b.WriteString("untagged-")
	}

	b.WriteString(tag)

	if !isClean {
		b.WriteString("-dirty")
	}

	if extra != "" {
		b.WriteRune('-')
		b.WriteString(extra)
	}

	b.WriteString(fmt.Sprintf(" (%s, %s)", branch, commitShort(commit)))

	return b.String()
}

func commitShort(commitLong string) (commit string) {
	if commitLong == "" {
		return unknown
	}

	b := strings.Builder{}

	for i, r := range commitLong {
		b.WriteRune(r)

		if i >= 6 {
			break
		}
	}

	return b.String()
}
