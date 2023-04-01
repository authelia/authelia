package utils

import (
	"fmt"
	"runtime"
)

// ErrSliceSortAlphabetical is a helper type that can be used with sort.Sort to sort a slice of errors in alphabetical
// order. Usage is simple just do sort.Sort(ErrSliceSortAlphabetical([]error{})).
type ErrSliceSortAlphabetical []error

func (s ErrSliceSortAlphabetical) Len() int { return len(s) }

func (s ErrSliceSortAlphabetical) Less(i, j int) bool { return s[i].Error() < s[j].Error() }

func (s ErrSliceSortAlphabetical) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetExpectedErrTxt returns error text for expected errs.
func GetExpectedErrTxt(err string) string {
	switch runtime.GOOS {
	case windows:
		switch err {
		case "pathnotfound":
			return fmt.Sprintf(errFmtWindowsNotFound, "open", "path")
		case "statpathnotfound":
			return fmt.Sprintf(errFmtWindowsNotFound, "stat", "path")
		case "filenotfound":
			return fmt.Sprintf(errFmtWindowsNotFound, "open", "file")
		case "statfilenotfound":
			return fmt.Sprintf(errFmtWindowsNotFound, "stat", "file")
		case "isdir":
			return "read %s: The handle is invalid."
		}
	default:
		switch err {
		case "pathnotfound", "filenotfound":
			return fmt.Sprintf(errFmtLinuxNotFound, "open")
		case "statpathnotfound", "statfilenotfound":
			return fmt.Sprintf(errFmtLinuxNotFound, "stat")
		case "isdir":
			return "read %s: is a directory"
		}
	}
	return ""
}
