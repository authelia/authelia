package utils

import (
	"bytes"
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
	buf := bytes.NewBuffer(nil)

	states := strings.Split(state, " ")

	isClean := IsStringInSlice(clean, states)
	isTagged := IsStringInSlice(tagged, states)

	if isClean && isTagged {
		buf.WriteString(tag)

		if extra != "" {
			buf.WriteRune('-')
			buf.WriteString(extra)
		}

		return buf.String()
	}

	if isTagged && !isClean {
		buf.WriteString(tag)
		buf.WriteString("-dirty")

		return buf.String()
	}

	if !isTagged {
		buf.WriteString("untagged-")
	}

	buf.WriteString(tag)

	if !isClean {
		buf.WriteString("-dirty")
	}

	if extra != "" {
		buf.WriteRune('-')
		buf.WriteString(extra)
	}

	if Dev {
		buf.WriteString("-dev")
	}

	_, _ = fmt.Fprint(buf, " (", branch, ", ", commitShort(commit), ")")

	return buf.String()
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
