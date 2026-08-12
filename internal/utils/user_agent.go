package utils

import (
	"fmt"

	"github.com/avct/uasurfer"
)

func FormatVersion(version uasurfer.Version) string {
	if version.Major == 0 && version.Minor == 0 && version.Patch == 0 {
		return "Unknown"
	}

	if version.Patch == 0 {
		if version.Minor == 0 {
			return fmt.Sprintf("%d", version.Major)
		}

		return fmt.Sprintf("%d.%d", version.Major, version.Minor)
	}

	return fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)
}

func ParseUserAgent(userAgent string) *uasurfer.UserAgent {
	return uasurfer.Parse(userAgent)
}
