package suites

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/matryer/is"
	"github.com/poy/onpar"
)

type Headers struct {
	ForwardedEmail  string `json:"Remote-Email"`
	ForwardedGroups string `json:"Remote-Groups"`
	ForwardedName   string `json:"Remote-Name"`
	ForwardedUser   string `json:"Remote-User"`
}

type HeadersPayload struct {
	Headers Headers `json:"headers"`
}

func TestRunCustomHeadersScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	o := onpar.New()
	defer o.Run(t)

	o.Group("TestCustomHeadersScenario", func() {
		o.BeforeEach(func(t *testing.T) (*testing.T, RodSuite) {
			s := setupTest(t, "", false)
			return t, s
		})

		o.AfterEach(func(t *testing.T, s RodSuite) {
			teardownTest(s)
		})

		o.Spec("TestShouldNotForwardCustomHeaderForUnauthenticatedUser", func(t *testing.T, s RodSuite) {
			is := is.New(t)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			s.doVisit(s.Context(ctx), fmt.Sprintf("%s/headers", PublicBaseURL))

			body, err := s.Context(ctx).Element("body")
			is.NoErr(err)

			b, err := body.Text()
			is.NoErr(err)
			is.True(!strings.Contains(b, testUsername))
			is.True(!strings.Contains(b, "admins"))
			is.True(!strings.Contains(b, "John Doe"))
			is.True(!strings.Contains(b, "john.doe@authelia.com"))
		})

		o.Spec("TestShouldForwardCustomHeaderForAuthenticatedUser", func(t *testing.T, s RodSuite) {
			is := is.New(t)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), s.Page)
			}()

			expectedGroups := mapset.NewSetWith("dev", "admins")

			targetURL := fmt.Sprintf("%s/headers", PublicBaseURL)
			s.doLoginOneFactor(t, s.Context(ctx), testUsername, testPassword, false, targetURL)
			s.verifyIsPublic(t, s.Context(ctx))

			body, err := s.Context(ctx).Element("body")
			is.NoErr(err)

			content, err := body.Text()
			is.NoErr(err)

			payload := HeadersPayload{}
			if err := json.Unmarshal([]byte(content), &payload); err != nil {
				log.Panic(err)
			}

			groups := strings.Split(payload.Headers.ForwardedGroups, ",")
			actualGroups := mapset.NewSet()

			for _, group := range groups {
				actualGroups.Add(group)
			}

			if strings.Contains(payload.Headers.ForwardedUser, testUsername) && expectedGroups.Equal(actualGroups) &&
				strings.Contains(payload.Headers.ForwardedName, "John Doe") && strings.Contains(payload.Headers.ForwardedEmail, "john.doe@authelia.com") {
				err = nil
			} else {
				err = fmt.Errorf("headers do not include user information")
			}
			is.NoErr(err)
		})
	})
}
