package suites

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/require"
	"github.com/ysmood/gson"
)

const notificationRecorder = `(binding) => {
	const seen = new Set();

	const report = (node) => {
		if (!node || node.nodeType !== 1) {
			return;
		}

		const matches = node.matches('.notification') ? [node] : Array.from(node.querySelectorAll('.notification'));

		for (const element of matches) {
			const text = element.textContent || '';

			if (text && !seen.has(text)) {
				seen.add(text);
				window[binding](text);
			}
		}
	};

	document.querySelectorAll('.notification').forEach(report);

	new MutationObserver((records) => {
		for (const record of records) {
			record.addedNodes.forEach(report);
		}
	}).observe(document.documentElement, {childList: true, subtree: true});
}`

func (rs *RodSession) verifyNotificationDisplayed(t *testing.T, page *rod.Page, message string) {
	el, err := page.ElementR(".notification", message)

	require.NoError(t, err)
	require.NotNil(t, el)
}

func (rs *RodSession) verifyNotificationDisplayedDuring(t *testing.T, page *rod.Page, message string, action func()) {
	// Recording from before the action runs means a notification torn down by a redirect the action
	// itself triggers is still observed; looking for it afterwards races that navigation.
	var (
		mutex sync.Mutex
		seen  []string
	)

	stop, err := page.Expose(notificationBinding, func(value gson.JSON) (any, error) {
		mutex.Lock()
		defer mutex.Unlock()

		seen = append(seen, value.Str())

		return nil, nil
	})

	require.NoError(t, err)

	defer func() {
		_ = stop()
	}()

	_, err = page.Eval(notificationRecorder, notificationBinding)

	require.NoError(t, err)

	action()

	ctx := page.GetContext()

	for {
		mutex.Lock()

		for _, text := range seen {
			if strings.Contains(text, message) {
				mutex.Unlock()

				return
			}
		}

		observed := append([]string(nil), seen...)

		mutex.Unlock()

		select {
		case <-ctx.Done():
			require.Failf(t, "Notification was not displayed",
				"expected a notification containing '%s', observed %v", message, observed)

			return
		case <-time.After(notificationPollInterval):
		}
	}
}
