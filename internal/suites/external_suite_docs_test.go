//go:build externalsuites
// +build externalsuites

package suites

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/html"
	"golang.org/x/sync/errgroup"
)

type DocsSuite struct {
	*RodSuite

	baseURL   string
	devServer *DevServer
	timeout   time.Duration
}

func NewDocsSuite() *DocsSuite {
	return &DocsSuite{
		RodSuite: NewRodSuite(externalSuiteNameDocs),
	}
}

func TestDocsSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping external suite in short mode")
	}

	suite.Run(t, NewDocsSuite())
}

func (s *DocsSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
	s.timeout = 10 * time.Second

	repoRoot, err := findRepoRoot()
	require.NoError(s.T(), err)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	srv, err := StartDevServer(ctx, repoRoot, HugoDocsDevServer, nil, func(early *DevServer) {
		globalDevServer.Store(early)
	})
	require.NoError(s.T(), err)

	s.devServer = srv
	s.baseURL = srv.BaseURL()

	browser, err := NewRodSession(RodSessionWithoutDevtools())
	require.NoError(s.T(), err)
	s.RodSession = browser
}

func (s *DocsSuite) TearDownSuite() {
	if s.RodSession != nil {
		if err := s.Stop(); err != nil {
			s.T().Logf("error stopping rod session: %v", err)
		}
	}

	if s.devServer != nil {
		if err := s.devServer.Stop(); err != nil {
			s.T().Logf("error stopping %s dev server: %v", s.devServer.Name(), err)
		}
	}

	globalDevServer.Store(nil)
}

func (s *DocsSuite) docsURL(path string) string {
	return s.baseURL + path
}

func (s *DocsSuite) httpFetch(ctx context.Context, pathOrURL string) (*http.Response, []byte) {
	client := &http.Client{Timeout: s.timeout}

	url := pathOrURL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = s.docsURL(pathOrURL)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(s.T(), err)

	resp, err := client.Do(req)
	require.NoError(s.T(), err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)

	return resp, body
}

func (s *DocsSuite) TestHomepageVisualSnapshot() {
	page := s.doCreateTab(s.T(), "about:blank")
	defer page.MustClose()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), page)
	}()

	page = page.Context(ctx)

	s.SetColorScheme(s.T(), page, "dark")
	page.MustSetViewport(1280, 800, 1, false)

	require.NoError(s.T(), page.Navigate(s.docsURL("/")))
	require.NoError(s.T(), page.WaitLoad())

	s.WaitElementLocatedByClassName(s.T(), page, "navbar")
	s.WaitForVisualStable(s.T(), page)

	screenshot := s.FullPageScreenshot(s.T(), page)

	repoRoot, err := findRepoRoot()
	require.NoError(s.T(), err)

	AssertVisualSnapshot(s.T(), repoRoot, "docs_homepage_snapshot.png", screenshot, VisualSnapshotTolerance(1.0))
}

func (s *DocsSuite) TestHomepageRendersAndSearch() {
	page := s.doCreateTab(s.T(), s.docsURL("/"))
	defer page.MustClose()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), page)
	}()

	page = page.Context(ctx)

	page.MustSetViewport(1280, 800, 1, false)

	s.WaitElementLocatedByClassName(s.T(), page, "navbar")

	s.WaitElementLocatedByID(s.T(), page, "searchToggleDesktop").MustClick()

	input := s.WaitElementLocatedByID(s.T(), page, "docsearch-input")
	input.MustWaitVisible()
	input.MustInput("OpenID Connect")

	hit := s.WaitElementLocatedBySelector(s.T(), page, ".DocSearch-Hit a")

	hitHrefPtr := hit.MustAttribute("href")
	require.NotNil(s.T(), hitHrefPtr, "expected the first search hit to have an href attribute")

	hitHref := *hitHrefPtr
	require.Contains(s.T(), strings.ToLower(hitHref), "openid", "expected first search hit to be OIDC-related, got %s", hitHref)

	homeURL := page.MustInfo().URL

	destURL := s.DoAndWaitForNavigation(ctx, page, func() {
		hit.MustClick()
	})

	require.NotEqual(s.T(), homeURL, destURL, "expected URL to change after clicking the search hit")
	require.Contains(s.T(), strings.ToLower(destURL), "openid", "expected to navigate to an OIDC page, got %q", destURL)

	heading := s.WaitElementLocatedBySelector(s.T(), page, "main h1, article h1, h1.page-title, h1")
	headingText := strings.ToLower(heading.MustText())
	require.NotEmpty(s.T(), headingText, "expected destination page to have a non-empty h1 at %s", destURL)
	require.True(s.T(),
		strings.Contains(headingText, "openid") || strings.Contains(headingText, "oidc"),
		"expected destination h1 to mention OpenID/OIDC at %s, got %q", destURL, headingText,
	)
}

func (s *DocsSuite) TestSupportedProxiesTable() {
	page := s.doCreateTab(s.T(), s.docsURL("/overview/prologue/supported-proxies/"))
	defer page.MustClose()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), page)
	}()

	page = page.Context(ctx)

	s.WaitElementLocatedBySelector(s.T(), page, "table")

	body := s.WaitElementLocatedBySelector(s.T(), page, "body").MustText()
	for _, proxy := range []string{"Traefik", "NGINX", "HAProxy"} {
		require.Contains(s.T(), body, proxy, "expected proxy %q in the support matrix", proxy)
	}

	s.WaitElementLocatedByClassName(s.T(), page, "icon-support-full")
	s.WaitElementLocatedByClassName(s.T(), page, "icon-support-unknown")
}

func (s *DocsSuite) TestOpenIDConnectIntroductionImages() {
	page := s.doCreateTab(s.T(), s.docsURL("/integration/openid-connect/introduction/"))
	defer page.MustClose()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), page)
	}()

	page = page.Context(ctx)

	// {{< figure >}} content-hashes the JPG through Hugo's asset pipeline, so match by
	// substring rather than literal path.
	for _, img := range []struct {
		selector    string
		contentType string
	}{
		{`img[src*="oid-certification"]`, "image/jpeg"},
		{`img[src$="/images/oid-map.png"]`, "image/png"},
	} {
		el := s.WaitElementLocatedBySelector(s.T(), page, img.selector)

		srcPtr := el.MustAttribute("src")
		require.NotNil(s.T(), srcPtr, "expected %s to have a src attribute", img.selector)

		resp, _ := s.httpFetch(ctx, *srcPtr)
		require.Equal(s.T(), http.StatusOK, resp.StatusCode, "expected 200 fetching %s", *srcPtr)
		require.True(s.T(), strings.HasPrefix(resp.Header.Get("Content-Type"), img.contentType), "expected %s for %s, got %s", img.contentType, *srcPtr, resp.Header.Get("Content-Type"))
	}
}

func (s *DocsSuite) TestStaticSvgAssetServed() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const svgPath = "/svgs/logos/jetbrains.svg"

	resp, body := s.httpFetch(ctx, svgPath)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode, "expected 200 fetching %s", svgPath)
	require.True(s.T(), strings.HasPrefix(resp.Header.Get("Content-Type"), "image/svg+xml"), "expected image/svg+xml for %s, got %s", svgPath, resp.Header.Get("Content-Type"))
	require.NotEmpty(s.T(), body, "expected non-empty body for %s", svgPath)
	require.True(s.T(), strings.Contains(string(body), "<svg"), "expected <svg root tag in body of %s", svgPath)
}

func (s *DocsSuite) TestOpenIDConnectProviderShortcodes() {
	page := s.doCreateTab(s.T(), s.docsURL("/configuration/identity-providers/openid-connect/provider/"))
	defer page.MustClose()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), page)
	}()

	page = page.Context(ctx)

	s.WaitElementLocatedByID(s.T(), page, "site-variables-toggle")
	s.WaitElementLocatedByID(s.T(), page, "site-variables-modal")
	s.WaitElementLocatedByClassName(s.T(), page, "callout-caution")

	span := s.WaitElementLocatedByClassName(s.T(), page, "site-variable-domain")

	text := span.MustText()
	require.Contains(s.T(), text, "example.com", "expected example.com nojs default in span.site-variable-domain, got %q", text)
}

const tableScrollWrapperClass = "table-responsive"

const tableOverflowJS = `() => {
	const doc = document.documentElement;

	const scroller = (el) => {
		for (let node = el.parentElement; node; node = node.parentElement) {
			const overflowX = getComputedStyle(node).overflowX;

			if (overflowX === 'auto' || overflowX === 'scroll' || overflowX === 'hidden') {
				return node;
			}
		}

		return null;
	};

	return {
		clientWidth: doc.clientWidth,
		tables: [...document.querySelectorAll('table')].map((table) => {
			const box = scroller(table);

			return {
				scrolls: box !== null,
				right: Math.round((box || table).getBoundingClientRect().right),
				width: Math.round(table.getBoundingClientRect().width),
			};
		}),
	};
}`

func (s *DocsSuite) TestTablesAreWrappedInScrollContainers() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	_, index := s.httpFetch(ctx, "/search-index.json")

	var entries []struct {
		Permalink string `json:"permalink"`
	}

	require.NoError(s.T(), json.Unmarshal(index, &entries))
	require.NotEmpty(s.T(), entries, "expected the search index to list at least one page")

	var (
		mutex     sync.Mutex
		unwrapped []string
		tables    int
	)

	client := &http.Client{Timeout: s.timeout}

	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(8)

	for _, entry := range entries {
		permalink := entry.Permalink

		group.Go(func() error {
			body, err := fetchPage(groupCtx, client, s.docsURL(permalink))
			if err != nil {
				return err
			}

			doc, err := html.Parse(bytes.NewReader(body))
			if err != nil {
				return fmt.Errorf("error parsing %s: %w", permalink, err)
			}

			found, bare := countTablesWithoutScrollWrapper(doc)

			mutex.Lock()

			tables += found

			if bare != 0 {
				unwrapped = append(unwrapped, fmt.Sprintf("%s (%d)", permalink, bare))
			}

			mutex.Unlock()

			return nil
		})
	}

	require.NoError(s.T(), group.Wait())

	require.NotZero(s.T(), tables, "expected the documentation to contain tables, the check is meaningless otherwise")
	require.Empty(s.T(), unwrapped, "expected every table to be wrapped in a .%s container, %d were not: %s", tableScrollWrapperClass, len(unwrapped), strings.Join(unwrapped, ", "))
}

func (s *DocsSuite) TestWideTablesDoNotOverflowViewport() {
	for _, path := range []string{
		// The page from https://github.com/authelia/authelia/discussions/12831.
		"/overview/security/measures/",
		"/configuration/methods/environment/",
		"/configuration/prologue/common/",
		"/integration/openid-connect/introduction/",
		// Renders its table as raw HTML rather than Markdown.
		"/reference/integrations/time-based-one-time-password-apps/",
	} {
		for _, viewport := range []struct {
			width, height int
		}{
			{1280, 800},
			{390, 844},
		} {
			s.Run(fmt.Sprintf("%s/%dx%d", strings.Trim(path, "/"), viewport.width, viewport.height), func() {
				page := s.doCreateTab(s.T(), s.docsURL(path))
				defer page.MustClose()

				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

				defer func() {
					cancel()
					s.collectScreenshot(ctx.Err(), page)
				}()

				page = page.Context(ctx)

				page.MustSetViewport(viewport.width, viewport.height, 1, false)

				s.WaitElementLocatedBySelector(s.T(), page, "table")

				result, err := page.Eval(tableOverflowJS)
				require.NoError(s.T(), err)

				clientWidth := result.Value.Get("clientWidth").Int()
				tables := result.Value.Get("tables").Arr()

				require.NotEmpty(s.T(), tables, "expected at least one table on %s", path)

				for i, table := range tables {
					require.True(s.T(), table.Get("scrolls").Bool(), "expected table %d on %s to sit inside a horizontal scroll container", i, path)
					require.LessOrEqual(s.T(), table.Get("right").Int(), clientWidth, "expected the scroll container of table %d on %s (table width %d) not to reach past the %dpx document edge", i, path, table.Get("width").Int(), clientWidth)
				}
			})
		}
	}
}

func fetchPage(ctx context.Context, client *http.Client, url string) (body []byte, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected 200 fetching %s, got %d", url, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func countTablesWithoutScrollWrapper(node *html.Node) (tables, bare int) {
	var walk func(node *html.Node, wrapped bool)

	walk = func(node *html.Node, wrapped bool) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "table":
				tables++

				if !wrapped {
					bare++
				}
			default:
				for _, attribute := range node.Attr {
					if attribute.Key == "class" && slices.Contains(strings.Fields(attribute.Val), tableScrollWrapperClass) {
						wrapped = true

						break
					}
				}
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child, wrapped)
		}
	}

	walk(node, false)

	return tables, bare
}

const documentOverflowJS = `() => {
	const doc = document.documentElement;

	return doc.scrollWidth - doc.clientWidth;
}`

func (s *DocsSuite) TestNavbarDoesNotOverflowViewport() {
	for _, width := range []int{1440, 1280, 1200, 1100, 1024, 992, 768, 390} {
		s.Run(fmt.Sprintf("%dx800", width), func() {
			page := s.doCreateTab(s.T(), s.docsURL("/"))
			defer page.MustClose()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), page)
			}()

			page = page.Context(ctx)

			page.MustSetViewport(width, 800, 1, false)

			s.WaitElementLocatedByClassName(s.T(), page, "navbar")

			result, err := page.Eval(documentOverflowJS)
			require.NoError(s.T(), err)

			require.Zero(s.T(), result.Value.Int(), "expected the homepage not to scroll horizontally at %dpx wide", width)
		})
	}
}

func (s *DocsSuite) TestContentDoesNotOverflowViewport() {
	for _, path := range []string{
		// Long configuration keys and environment variable names in headings and prose.
		"/configuration/identity-providers/openid-connect/clients/",
		"/configuration/identity-providers/openid-connect/provider/",
		"/configuration/methods/secrets/",
		"/configuration/miscellaneous/server-endpoint-rate-limits/",
		"/configuration/second-factor/webauthn/",
		// Runs of footnote back references joined by non-breaking spaces.
		"/blog/technical-openid-connect-1.0-nuances/",
	} {
		s.Run(strings.Trim(path, "/"), func() {
			page := s.doCreateTab(s.T(), s.docsURL(path))
			defer page.MustClose()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			defer func() {
				cancel()
				s.collectScreenshot(ctx.Err(), page)
			}()

			page = page.Context(ctx)

			page.MustSetViewport(390, 844, 1, false)

			s.WaitElementLocatedByClassName(s.T(), page, "content")

			result, err := page.Eval(documentOverflowJS)
			require.NoError(s.T(), err)

			require.Zero(s.T(), result.Value.Int(), "expected %s not to scroll horizontally at 390px wide", path)
		})
	}
}
