//go:build externalsuites
// +build externalsuites

package suites

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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
		globalDevServer = early
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

	globalDevServer = nil
}

func (s *DocsSuite) docsURL(path string) string {
	return s.baseURL + path
}

func (s *DocsSuite) httpFetch(ctx context.Context, path string) (*http.Response, []byte) {
	client := &http.Client{Timeout: s.timeout}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.docsURL(path), nil)
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

	AssertVisualSnapshot(s.T(), repoRoot, "docs_homepage_snapshot.png", screenshot, 1.0)
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
	s.WaitElementLocatedBySelector(s.T(), page, `img[src*="oid-certification"]`)
	s.WaitElementLocatedBySelector(s.T(), page, `img[src$="/images/oid-map.png"]`)

	for _, img := range []struct {
		path        string
		contentType string
	}{
		{"/images/oid-certification.jpg", "image/jpeg"},
		{"/images/oid-map.png", "image/png"},
	} {
		resp, _ := s.httpFetch(ctx, img.path)
		require.Equal(s.T(), http.StatusOK, resp.StatusCode, "expected 200 fetching %s", img.path)
		require.True(s.T(), strings.HasPrefix(resp.Header.Get("Content-Type"), img.contentType), "expected %s for %s, got %s", img.contentType, img.path, resp.Header.Get("Content-Type"))
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
