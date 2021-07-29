package utils

import (
	"fmt"
	"strconv"
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
// BuildTag i.e. v1.0.0. If dirty and tagged are present it returns <BuildTag>-dirty. Otherwise the following is the
// format: untagged-<BuildTag>-dirty-<BuildExtra> (<BuildBranch>, <BuildCommit>).
//
func Version() (versionString string) {
	return version(BuildTag, BuildState, BuildCommit, BuildBranch, BuildExtra)
}

func version(tag, state, commit, branch, extra string) (version string) {
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

// SemanticVersion represents a semantic 2.0 version.
type SemanticVersion struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Metadata   []string
}

// String is a function to provide a nice representation of a SemanticVersion.
func (v SemanticVersion) String() (value string) {
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch))

	if v.PreRelease != "" {
		builder.WriteString("-")
		builder.WriteString(v.PreRelease)
	}

	if len(v.Metadata) != 0 {
		builder.WriteString("+")
		builder.WriteString(strings.Join(v.Metadata, "."))
	}

	return builder.String()
}

// Equals returns true if this SemanticVersion is equal to the provided SemanticVersion.
func (v SemanticVersion) Equals(version SemanticVersion) (equals bool) {
	return v.Major == version.Major && v.Minor == version.Minor && v.Patch == version.Patch && v.PreRelease == version.PreRelease
}

// Greater returns true if this SemanticVersion is greater than the provided SemanticVersion.
func (v SemanticVersion) Greater(version SemanticVersion) (greater bool) {
	if v.Major > version.Major || v.Minor > version.Minor || v.Patch > version.Patch {
		return true
	}

	if v.PreRelease == "" && version.PreRelease != "" {
		return true
	}

	if v.PreRelease != "" && version.PreRelease != "" {
		if strings.Compare(v.PreRelease, version.PreRelease) == 1 {
			return true
		}
	}

	return false
}

// NewSemanticVersion creates a SemanticVersion from a string.
func NewSemanticVersion(input string) (version SemanticVersion) {
	submatch := reSemanticVersion.FindStringSubmatch(input)

	for i, name := range reSemanticVersion.SubexpNames() {
		switch name {
		case "Major":
			version.Major, _ = strconv.Atoi(submatch[i])
		case "Minor":
			version.Minor, _ = strconv.Atoi(submatch[i])
		case "Patch":
			version.Patch, _ = strconv.Atoi(submatch[i])
		case "PreRelease":
			version.PreRelease = submatch[i]
		case "Metadata":
			if submatch[i] != "" {
				version.Metadata = strings.Split(submatch[i], ".")
			}
		}
	}

	return version
}
