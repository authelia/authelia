package utils

import "runtime"

// GetExpectedErrTxt returns error text for expected errs.
func GetExpectedErrTxt(err string) string {
	switch err {
	case "pathnotfound":
		switch runtime.GOOS {
		case windows:
			return "open %s: The system cannot find the path specified."
		default:
			return errFmtLinuxNotFound
		}
	case "filenotfound":
		switch runtime.GOOS {
		case windows:
			return "open %s: The system cannot find the file specified."
		default:
			return errFmtLinuxNotFound
		}
	case "yamlisdir":
		switch runtime.GOOS {
		case windows:
			return "read %s: The handle is invalid."
		default:
			return "read %s: is a directory"
		}
	}

	return ""
}
