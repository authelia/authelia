package model

import (
	"fmt"
	"strconv"
	"strings"
)

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
		case semverRegexpGroupPreRelease, "Metadata":
			val := make([]string, 0)
			if submatch[i] != "" {
				val = strings.Split(submatch[i], ".")
			}

			if name == semverRegexpGroupPreRelease {
				version.PreRelease = val
			} else {
				version.Metadata = val
			}
		}
	}

	return version
}

// SemanticVersion represents a semantic 2.0 version.
type SemanticVersion struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease []string
	Metadata   []string
}

// String is a function to provide a nice representation of a SemanticVersion.
func (v SemanticVersion) String() (value string) {
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch))

	if len(v.PreRelease) != 0 {
		builder.WriteString("-")
		builder.WriteString(strings.Join(v.PreRelease, "."))
	}

	if len(v.Metadata) != 0 {
		builder.WriteString("+")
		builder.WriteString(strings.Join(v.Metadata, "."))
	}

	return builder.String()
}

// Equal returns true if this SemanticVersion is equal to the provided SemanticVersion.
func (v SemanticVersion) Equal(version SemanticVersion) (equals bool) {
	return v.Major == version.Major && v.Minor == version.Minor && v.Patch == version.Patch
}

// GreaterThan returns true if this SemanticVersion is greater than the provided SemanticVersion.
func (v SemanticVersion) GreaterThan(version SemanticVersion) (gt bool) {
	if v.Major < version.Major {
		return false
	}

	if v.Major == version.Major && v.Minor < version.Minor {
		return false
	}

	if v.Major == version.Major && v.Minor == version.Minor && v.Patch < version.Patch {
		return false
	}

	return true
}

// GreaterThanOrEqual returns true if this SemanticVersion is greater than the provided SemanticVersion.
func (v SemanticVersion) GreaterThanOrEqual(version SemanticVersion) (ge bool) {
	return v.GreaterThan(version) || v.Equal(version)
}
