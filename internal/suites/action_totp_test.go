package suites

import (
	"net/url"
	"testing"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/stretchr/testify/require"
)

// TestOTPEntered covers the check that the passcode reached the field, which has to tell a field that
// never received it apart from one taken away because it did.
func TestOTPEntered(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	path, err := GetBrowserPath()
	require.NoError(t, err)

	l := launcher.New().Bin(path).Headless(true)
	defer l.Cleanup()

	browser := rod.New().ControlURL(l.MustLaunch()).MustConnect()
	defer func() { _ = browser.Close() }()

	page := browser.MustPage("data:text/html," + url.PathEscape(`<div id="host"></div>`))

	set := func(html string) {
		_, err := page.Eval(`(html) => { document.getElementById('host').innerHTML = html; }`, html)
		require.NoError(t, err)
	}
	entered := func() bool {
		r, err := page.Eval(otpEntered, "123456")
		require.NoError(t, err)

		return r.Value.Bool()
	}

	slots := func(vals ...string) string {
		out := `<span id="otp-input">`
		for _, v := range vals {
			out += `<input value="` + v + `">`
		}

		out += `<input aria-hidden="true" value="123456"></span>`

		return out
	}

	set(slots("1", "2", "3", "4", "5", "6"))
	require.True(t, entered(), "a filled field is entered")

	set(slots("", "", "", "", "", ""))
	require.False(t, entered(), "an empty field is NOT entered - this is the regression guard")

	set(slots("1", "2", "3", "", "", ""))
	require.False(t, entered(), "a partially filled field is not entered")

	set(slots("9", "9", "9", "9", "9", "9"))
	require.False(t, entered(), "the wrong passcode is not entered")

	set(``)
	require.True(t, entered(), "a field taken away after submission counts as entered")
}
