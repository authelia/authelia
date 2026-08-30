//go:build externalsuites
// +build externalsuites

package suites

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/cdp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TemplatesSuite struct {
	*RodSuite

	baseURL   string
	devServer *DevServer
	timeout   time.Duration
}

func NewTemplatesSuite() *TemplatesSuite {
	return &TemplatesSuite{
		RodSuite: NewRodSuite(externalSuiteNameTemplates),
	}
}

func TestTemplatesSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping external suite in short mode")
	}

	suite.Run(t, NewTemplatesSuite())
}

func (s *TemplatesSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
	s.timeout = 15 * time.Second

	repoRoot, err := findRepoRoot()
	require.NoError(s.T(), err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cfg := ReactEmailTemplatesDevServer
	cfg.StartTimeout = 2 * time.Minute

	srv, err := StartDevServer(ctx, repoRoot, cfg, nil, func(early *DevServer) {
		globalDevServer.Store(early)
	})
	require.NoError(s.T(), err)

	s.devServer = srv
	s.baseURL = srv.BaseURL()

	browser, err := NewRodSession(RodSessionWithoutDevtools())
	require.NoError(s.T(), err)
	s.RodSession = browser
}

func (s *TemplatesSuite) TearDownSuite() {
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

func (s *TemplatesSuite) templatesURL(path string) string {
	return s.baseURL + path
}

func isStaleDocumentError(err error) bool {
	var cdpErr *cdp.Error

	return errors.As(err, &cdpErr) && strings.Contains(cdpErr.Message, "does not belong to the document")
}

func (s *TemplatesSuite) previewText(outer *rod.Page, readySelector, selector string) string {
	// react-email swaps the iframe document when it finishes bundling, so there is no moment at which
	// descending is guaranteed safe. Rather than trying to pick one, the frame is resolved again
	// whenever a node turns out to belong to a replaced document.
	var (
		text string
		err  error
	)

	for i := 0; i < waitElementsAttempts; i++ {
		var element *rod.Element

		if element, err = s.openPreviewFrame(outer, readySelector).Element(selector); err != nil {
			if isStaleDocumentError(err) {
				continue
			}

			break
		}

		if text, err = element.Text(); err == nil {
			return text
		}

		if !isStaleDocumentError(err) {
			break
		}
	}

	require.NoError(s.T(), err, "failed to read '%s' from the preview iframe", selector)

	return text
}

func (s *TemplatesSuite) openPreviewFrame(outer *rod.Page, readySelector string) *rod.Page {
	s.WaitElementLocatedBySelector(s.T(), outer, "iframe")

	// readySelector is only present once react-email has finished populating the preview iframe, so
	// waiting on it is what stops a partial or empty srcdoc being captured.
	outer.MustWait(`(sel) => {
		const f = document.querySelector('iframe');
		return !!(f && f.contentDocument && f.contentDocument.querySelector(sel));
	}`, readySelector)

	iframeEl := s.WaitElementLocatedBySelector(s.T(), outer, "iframe")

	frame, err := iframeEl.Frame()
	require.NoError(s.T(), err, "failed to descend into preview iframe")

	return frame
}

func (s *TemplatesSuite) TestPreviewIndexListsTemplates() {
	page := s.doCreateTab(s.T(), s.templatesURL("/"))
	defer page.MustClose()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), page)
	}()

	page = page.Context(ctx)

	for _, slug := range []string{"IdentityVerificationOTC", "IdentityVerificationJWT", "Event"} {
		s.WaitElementLocatedBySelector(s.T(), page, `a[href="/preview/`+slug+`"]`)
	}
}

func (s *TemplatesSuite) TestIdentityVerificationOTCRenders() {
	outer := s.doCreateTab(s.T(), s.templatesURL("/preview/IdentityVerificationOTC"))
	defer outer.MustClose()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), outer)
	}()

	outer = outer.Context(ctx)

	require.Contains(s.T(), s.previewText(outer, "#one-time-code", "#one-time-code"), "ABC123", "expected one-time code to render the PreviewProps value")

	s.previewText(outer, "#one-time-code", "#link-revoke")

	body := s.previewText(outer, "#one-time-code", "body")
	require.Contains(s.T(), strings.ToLower(body), "one-time code")
}

func (s *TemplatesSuite) TestIdentityVerificationJWTRenders() {
	outer := s.doCreateTab(s.T(), s.templatesURL("/preview/IdentityVerificationJWT"))
	defer outer.MustClose()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), outer)
	}()

	outer = outer.Context(ctx)

	s.previewText(outer, "#link", "#link")
	s.previewText(outer, "#link", "#link-revoke")

	body := s.previewText(outer, "#link", "body")
	require.Contains(s.T(), strings.ToLower(body), "one-time link")
}

func (s *TemplatesSuite) TestEventRenders() {
	outer := s.doCreateTab(s.T(), s.templatesURL("/preview/Event"))
	defer outer.MustClose()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), outer)
	}()

	outer = outer.Context(ctx)

	body := s.previewText(outer, "strong", "body")
	for _, needle := range []string{
		"Second Factor Method Added",
		"Example Detail",
		"Example Value",
		"Second Factor Method",
	} {
		require.Contains(s.T(), body, needle, "expected Event preview body to contain %q", needle)
	}
}

func (s *TemplatesSuite) injectEmbeddedFont(repoRoot, srcdoc string) string {
	// Forcing one font through the document means visual snapshots rasterize from the same outlines
	// on every host.
	fontPath := filepath.Join(repoRoot, "internal", "suites", "testdata", "fonts", "LiberationSans-Regular.ttf")

	fontBytes, err := os.ReadFile(fontPath)
	require.NoError(s.T(), err)

	fontB64 := base64.StdEncoding.EncodeToString(fontBytes)

	style := `<style>
@font-face {
	font-family: 'SnapshotSans';
	src: url(data:font/ttf;base64,` + fontB64 + `) format('truetype');
	font-weight: normal;
	font-style: normal;
}
html, body, * {
	font-family: 'SnapshotSans', sans-serif !important;
}
</style>`

	if i := strings.Index(srcdoc, "</head>"); i != -1 {
		return srcdoc[:i] + style + srcdoc[i:]
	}

	return style + srcdoc
}

func (s *TemplatesSuite) runTemplateSnapshot(slug, readySelector, snapshotName string) {
	outer := s.doCreateTab(s.T(), s.templatesURL("/preview/"+slug))
	defer outer.MustClose()

	clean := s.doCreateTab(s.T(), "about:blank")
	defer clean.MustClose()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), clean)
	}()

	outer = outer.Context(ctx)
	clean = clean.Context(ctx)

	s.WaitElementLocatedBySelector(s.T(), outer, "iframe")

	outer.MustWait(`(sel) => {
		const f = document.querySelector('iframe');
		return !!(f && f.contentDocument && f.contentDocument.querySelector(sel));
	}`, readySelector)

	iframeEl := s.WaitElementLocatedBySelector(s.T(), outer, "iframe")

	srcdocAttr := iframeEl.MustAttribute("srcdoc")
	require.NotNil(s.T(), srcdocAttr, "expected preview iframe to have a srcdoc attribute")

	repoRoot, err := findRepoRoot()
	require.NoError(s.T(), err)

	clean.MustSetViewport(800, 1200, 1, false)

	require.NoError(s.T(), clean.SetDocumentContent(s.injectEmbeddedFont(repoRoot, *srcdocAttr)))

	s.WaitForVisualStable(s.T(), clean)

	screenshot := s.FullPageScreenshot(s.T(), clean)

	AssertVisualSnapshot(s.T(), repoRoot, snapshotName, screenshot, VisualSnapshotTolerance(0))
}

func (s *TemplatesSuite) TestIdentityVerificationOTCVisualSnapshot() {
	s.runTemplateSnapshot("IdentityVerificationOTC", "#one-time-code", "templates_identity_verification_otc_snapshot.png")
}

func (s *TemplatesSuite) TestIdentityVerificationJWTVisualSnapshot() {
	s.runTemplateSnapshot("IdentityVerificationJWT", "#link", "templates_identity_verification_jwt_snapshot.png")
}

func (s *TemplatesSuite) TestEventVisualSnapshot() {
	s.runTemplateSnapshot("Event", "strong", "templates_event_snapshot.png")
}
