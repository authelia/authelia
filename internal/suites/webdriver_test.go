package suites

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/stretchr/testify/require"
)

const elementStatesHTML = `<style>body{margin:0}button{width:120px;height:40px}</style>
<button id="ok">ok</button>
<button id="disabled" disabled style="pointer-events:none">disabled</button>
<div style="position:relative;width:120px;height:40px"><button id="covered">covered</button>
<div id="veil" style="position:absolute;inset:0"></div></div>
<button id="offscreen" style="position:absolute;left:-9999px">off</button>
<input id="field" type="text">`

// TestElementActions covers the retry and reporting the suites rely on for every click and every keypress.
func TestElementActions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	path, err := GetBrowserPath()
	require.NoError(t, err)

	l := launcher.New().Bin(path).Headless(true)
	defer l.Cleanup()

	browser := rod.New().ControlURL(l.MustLaunch()).MustConnect()
	defer func() { _ = browser.Close() }()

	page := browser.MustPage("data:text/html," + url.PathEscape(elementStatesHTML))

	click := func(element *rod.Element) error { return element.Click("left", 1) }

	attempt := func(selector string, action func(*rod.Element) error) (time.Duration, error) {
		ctx, cancel := context.WithTimeout(page.GetContext(), time.Second*2)
		defer cancel()

		started := time.Now()
		err := retryElementAction(page.Context(ctx), selector, action)

		return time.Since(started), err
	}

	t.Run("ShouldClickAnActionableElement", func(t *testing.T) {
		_, err := attempt("#ok", click)
		require.NoError(t, err)
	})

	t.Run("ShouldConfirmTheValueItTyped", func(t *testing.T) {
		_, err := attempt("#field", func(element *rod.Element) error {
			if err := element.SelectAllText(); err != nil {
				return err
			}

			if err := element.Type((&RodSession{}).toInputs("typed")...); err != nil {
				return err
			}

			property, err := element.Property("value")
			if err != nil {
				return err
			}

			if property.Str() != "typed" {
				return errElementValueNotSet
			}

			return nil
		})

		require.NoError(t, err)
	})

	// Every one of these is a state the migrated components put an element in: disabled while a request
	// is in flight, covered by a portalled overlay, or moved out of the viewport by a transition.
	t.Run("ShouldGiveUpWithinItsContext", func(t *testing.T) {
		for _, tc := range []struct{ name, selector, state string }{
			{"Disabled", "#disabled", `"disabled":true`},
			{"Covered", "#covered", `"topmostAtCenter":"#veil"`},
			{"Offscreen", "#offscreen", `"rect":[-9999`},
		} {
			t.Run(tc.name, func(t *testing.T) {
				elapsed, err := attempt(tc.selector, click)

				require.Error(t, err)
				require.Less(t, elapsed, time.Second*5, "must stop when its context does")
				require.Contains(t, describeElement(page, tc.selector), tc.state)
			})
		}
	})

	t.Run("ShouldReportAnElementThatIsNotThere", func(t *testing.T) {
		require.Contains(t, describeElement(page, "#missing"), "not in the document")
	})
}
