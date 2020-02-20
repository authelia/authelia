package authorization

import "regexp"

func isPathMatching(path string, pathRegexps []string) bool {
	// If there is no regexp patterns, it means that we match any path.
	if len(pathRegexps) == 0 {
		return true
	}

	for _, pathRegexp := range pathRegexps {
		match, err := regexp.MatchString(pathRegexp, path)
		if err != nil {
			// TODO(c.michaud): make sure this is safe in advance to
			// avoid checking this case here.
			continue
		}

		if match {
			return true
		}
	}
	return false
}
