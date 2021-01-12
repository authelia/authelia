package authorization

import "regexp"

func isPathMatching(path string, pathRegexps []string) bool {
	// If there is no regexp patterns, it means that we match any path.
	if len(pathRegexps) == 0 {
		return true
	}

	for _, pathRegexp := range pathRegexps {
		match, _ := regexp.MatchString(pathRegexp, path)
		if match {
			return true
		}
	}

	return false
}
