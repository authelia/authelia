package utils

import (
	"testing"

	"github.com/avct/uasurfer"
	"github.com/stretchr/testify/assert"
)

func TestFormatVersion(t *testing.T) {
	testCases := []struct {
		name     string
		version  uasurfer.Version
		expected string
	}{
		{
			name:     "Zero version (unknown)",
			version:  uasurfer.Version{Major: 0, Minor: 0, Patch: 0},
			expected: "Unknown",
		},
		{
			name:     "Major version only",
			version:  uasurfer.Version{Major: 17, Minor: 0, Patch: 0},
			expected: "17",
		},
		{
			name:     "Major and minor versions",
			version:  uasurfer.Version{Major: 12, Minor: 3, Patch: 0},
			expected: "12.3",
		},
		{
			name:     "Full version",
			version:  uasurfer.Version{Major: 121, Minor: 5, Patch: 2345},
			expected: "121.5.2345",
		},
		{
			name:     "Zero patch with non-zero minor",
			version:  uasurfer.Version{Major: 9, Minor: 2, Patch: 0},
			expected: "9.2",
		},
		{
			name:     "Zero minor with non-zero patch (edge case)",
			version:  uasurfer.Version{Major: 8, Minor: 0, Patch: 1},
			expected: "8.0.1",
		},
		{
			name:     "Very large version numbers",
			version:  uasurfer.Version{Major: 999, Minor: 999, Patch: 9999},
			expected: "999.999.9999",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatVersion(tc.version)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseUserAgent(t *testing.T) {
	testCases := []struct {
		name           string
		userAgentStr   string
		expectedFields map[string]interface{}
	}{
		{
			name:         "Chrome on Windows",
			userAgentStr: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			expectedFields: map[string]interface{}{
				"BrowserName":  uasurfer.BrowserChrome,
				"BrowserMajor": 91,
				"OSName":       uasurfer.OSWindows,
				"OSMajor":      10,
				"DeviceType":   uasurfer.DeviceComputer,
			},
		},
		{
			name:         "Firefox on Linux",
			userAgentStr: "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0",
			expectedFields: map[string]interface{}{
				"BrowserName":  uasurfer.BrowserFirefox,
				"BrowserMajor": 89,
				"OSName":       uasurfer.OSLinux,
				"DeviceType":   uasurfer.DeviceComputer,
			},
		},
		{
			name:         "Safari on iPhone",
			userAgentStr: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Mobile/15E148 Safari/604.1",
			expectedFields: map[string]interface{}{
				"BrowserName":  uasurfer.BrowserSafari,
				"BrowserMajor": 14,
				"OSName":       uasurfer.OSiOS,
				"OSMajor":      14,
				"OSMinor":      6,
				"DeviceType":   uasurfer.DevicePhone,
			},
		},
		{
			name:         "Empty user agent",
			userAgentStr: "",
			expectedFields: map[string]interface{}{
				"BrowserName": uasurfer.BrowserUnknown,
				"OSName":      uasurfer.OSUnknown,
				"DeviceType":  uasurfer.DeviceUnknown,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseUserAgent(tc.userAgentStr)

			if val, ok := tc.expectedFields["BrowserName"]; ok {
				assert.Equal(t, val, result.Browser.Name)
			}

			if val, ok := tc.expectedFields["BrowserMajor"]; ok {
				assert.Equal(t, val, result.Browser.Version.Major)
			}

			if val, ok := tc.expectedFields["OSName"]; ok {
				assert.Equal(t, val, result.OS.Name)
			}

			if val, ok := tc.expectedFields["OSMajor"]; ok {
				assert.Equal(t, val, result.OS.Version.Major)
			}

			if val, ok := tc.expectedFields["OSMinor"]; ok {
				assert.Equal(t, val, result.OS.Version.Minor)
			}

			if val, ok := tc.expectedFields["DeviceType"]; ok {
				assert.Equal(t, val, result.DeviceType)
			}
		})
	}
}
