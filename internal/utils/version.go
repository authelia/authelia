package utils

import "strings"

// BuildTag is replaced by LDFLAGS at build time with the latest tag at or before the current commit.
var BuildTag = ""

// BuildStateTag is replaced by LDFLAGS at build time with `tagged` or `untagged` depending on if the commit is tagged.
var BuildStateTag = ""

// BuildStateExtra is replaced by LDFLAGS at build time with a blank string or `dirty` if the working tree is dirty.
var BuildStateExtra = ""

// BuildDate is replaced by LDFLAGS at build time with the date the build started.
var BuildDate = ""

// BuildCommit is replaced by LDFLAGS at build time with the current commit.
var BuildCommit = "unknown"

// BuildBranch is replaced by LDFLAGS at build time with the current branch.
var BuildBranch = ""

var versionLong = ""
var versionShort = ""

// VersionShort returns the short version.
//
// The format of the string is `<latest tag>-<short commit>-<extra>`. Where short commit and the hyphen are only present
// when the commit is not tagged, and the extra is only present if the extra is not blank. Extra is usually used to
// communicate if the working tree is dirty. Though it can be used by ports to include the port name.
//
func VersionShort() (version string) {
	if versionShort != "" {
		return versionShort
	}

	b := strings.Builder{}

	b.WriteString(BuildTag)

	if BuildStateTag != "tagged" {
		b.WriteRune('-')
		switch BuildCommit {
		case "uknown":
			b.WriteString("unknown")
		default:
			for i, r := range BuildCommit {
				b.WriteRune(r)
				if i >= 6 {
					break
				}
			}
		}
	}

	if BuildStateExtra != "" {
		b.WriteRune('-')
		b.WriteString(BuildStateExtra)
	}

	versionShort = b.String()

	return versionShort
}

// VersionLong returns the long version.
//
// The format of the string is `<latest tag>-untagged-<extra> (<branch>, <commit>, <date>)`. Where untagged and the
// hyphen are only present when the commit is not tagged, and the extra is only present if the extra is not blank. Extra
// is usually used to communicate if the working tree is dirty. Though it can be used by ports to include the port name.
// Branch is the name of the branch it was built from, commit is the long commit hash, and date is the time when the
// build started.
//
func VersionLong() (version string) {
	if versionLong != "" {
		return versionLong
	}

	b := strings.Builder{}

	b.WriteString(BuildTag)

	if BuildStateTag != "tagged" {
		b.WriteString("-untagged")
	}

	if BuildStateExtra != "" {
		b.WriteRune('-')
		b.WriteString(BuildStateExtra)
	}

	b.WriteString(" (")
	b.WriteString(BuildBranch)
	b.WriteString(", ")
	b.WriteString(BuildCommit)
	b.WriteString(", ")
	b.WriteString(BuildDate)
	b.WriteString(")")

	versionLong = b.String()

	return versionLong
}
