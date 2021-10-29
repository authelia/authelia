package suites

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-rod/rod"
	"github.com/matryer/is"
)

func (rs *RodSession) verifyBodyContains(t *testing.T, page *rod.Page, pattern string) {
	is := is.New(t)
	body, err := page.Element("body")
	is.NoErr(err)
	is.True(body != nil)

	text, err := body.Text()
	is.NoErr(err)
	is.True(text != "")

	if strings.Contains(text, pattern) {
		err = nil
	} else {
		err = fmt.Errorf("body does not contain pattern: %s", pattern)
	}

	is.NoErr(err)
}
