package model

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// NewSemanticVersion creates a SemanticVersion from a string.
func NewSemanticVersion(input string) (version *SemanticVersion, err error) {
	if !reSemanticVersion.MatchString(input) {
		return nil, fmt.Errorf("the input '%s' failed to match the semantic version pattern", input)
	}

	version = &SemanticVersion{}

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
			if submatch[i] == "" {
				continue
			}

			val := strings.Split(submatch[i], ".")

			if name == semverRegexpGroupPreRelease {
				version.PreRelease = val
			} else {
				version.Metadata = val
			}
		}
	}

	return version, nil
}

// SemanticVersion represents a semantic 2.0 version.
type SemanticVersion struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease []string
	Metadata   []string
}

// Copy the values for this SemanticVersion.
func (v SemanticVersion) Copy() SemanticVersion {
	return SemanticVersion{
		Major:      v.Major,
		Minor:      v.Minor,
		Patch:      v.Patch,
		PreRelease: v.PreRelease,
		Metadata:   v.Metadata,
	}
}

// IsStable returns true if the pre release and metadata values are empty and the major value is above 0.
func (v SemanticVersion) IsStable() bool {
	return v.IsAbsolute() && v.Major > 0
}

// IsAbsolute returns true if the pre release and metadata values are empty.
func (v SemanticVersion) IsAbsolute() bool {
	return len(v.PreRelease) == 0 && len(v.Metadata) == 0
}

// String is a function to provide a nice representation of a SemanticVersion.
func (v SemanticVersion) String() (value string) {
	buf := bytes.NewBuffer(nil)

	_, _ = fmt.Fprintf(buf, "%d.%d.%d", v.Major, v.Minor, v.Patch)

	if len(v.PreRelease) != 0 {
		buf.WriteString("-")
		buf.WriteString(strings.Join(v.PreRelease, "."))
	}

	if len(v.Metadata) != 0 {
		buf.WriteString("+")
		buf.WriteString(strings.Join(v.Metadata, "."))
	}

	return buf.String()
}

// Equal returns true if this SemanticVersion is equal to the provided SemanticVersion.
func (v SemanticVersion) Equal(version SemanticVersion) (equals bool) {
	return v.Major == version.Major && v.Minor == version.Minor && v.Patch == version.Patch
}

// GreaterThan returns true if this SemanticVersion is greater than the provided SemanticVersion.
func (v SemanticVersion) GreaterThan(version SemanticVersion) (gt bool) {
	if v.Major > version.Major {
		return true
	}

	if v.Major == version.Major && v.Minor > version.Minor {
		return true
	}

	if v.Major == version.Major && v.Minor == version.Minor && v.Patch > version.Patch {
		return true
	}

	return false
}

// LessThan returns true if this SemanticVersion is less than the provided SemanticVersion.
func (v SemanticVersion) LessThan(version SemanticVersion) (gt bool) {
	if v.Major < version.Major {
		return true
	}

	if v.Major == version.Major && v.Minor < version.Minor {
		return true
	}

	if v.Major == version.Major && v.Minor == version.Minor && v.Patch < version.Patch {
		return true
	}

	return false
}

// GreaterThanOrEqual returns true if this SemanticVersion is greater than or equal to the provided SemanticVersion.
func (v SemanticVersion) GreaterThanOrEqual(version SemanticVersion) (ge bool) {
	return v.Equal(version) || v.GreaterThan(version)
}

// LessThanOrEqual returns true if this SemanticVersion is less than or equal to the provided SemanticVersion.
func (v SemanticVersion) LessThanOrEqual(version SemanticVersion) (ge bool) {
	return v.Equal(version) || v.LessThan(version)
}

// NextMajor returns the next major SemanticVersion from this current SemanticVersion.
func (v SemanticVersion) NextMajor() (version SemanticVersion) {
	return SemanticVersion{Major: v.Major + 1}
}

// NextMinor returns the next minor SemanticVersion from this current SemanticVersion.
func (v SemanticVersion) NextMinor() (version SemanticVersion) {
	return SemanticVersion{Major: v.Major, Minor: v.Minor + 1}
}

// NextPatch returns the next patch SemanticVersion from this current SemanticVersion.
func (v SemanticVersion) NextPatch() (version SemanticVersion) {
	return SemanticVersion{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
}
