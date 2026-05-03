//go:build externalsuites
// +build externalsuites

package suites

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
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

// openPreviewFrame returns the inner page of the preview server's srcdoc iframe once
// react-email has rendered the element matching readySelector. Waiting on the final
// target selector (rather than any body content) avoids descending while react-email
// is still swapping documents, which would leave the frame handle pointing at a
// detached DOM and surface as "Node with given id does not belong to the document".
func (s *TemplatesSuite) openPreviewFrame(outer *rod.Page, readySelector string) *rod.Page {
	s.WaitElementLocatedBySelector(s.T(), outer, "iframe")

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

	frame := s.openPreviewFrame(outer, "#one-time-code")

	code := s.WaitElementLocatedByID(s.T(), frame, "one-time-code")
	require.Contains(s.T(), code.MustText(), "ABC123", "expected one-time code to render the PreviewProps value")

	s.WaitElementLocatedByID(s.T(), frame, "link-revoke")

	body := s.WaitElementLocatedBySelector(s.T(), frame, "body").MustText()
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

	frame := s.openPreviewFrame(outer, "#link")

	s.WaitElementLocatedByID(s.T(), frame, "link")
	s.WaitElementLocatedByID(s.T(), frame, "link-revoke")

	body := s.WaitElementLocatedBySelector(s.T(), frame, "body").MustText()
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

	frame := s.openPreviewFrame(outer, "strong")

	body := s.WaitElementLocatedBySelector(s.T(), frame, "body").MustText()
	for _, needle := range []string{
		"Second Factor Method Added",
		"Example Detail",
		"Example Value",
		"Second Factor Method",
	} {
		require.Contains(s.T(), body, needle, "expected Event preview body to contain %q", needle)
	}
}

// injectEmbeddedFont rewrites srcdoc to force Liberation Sans via a data-URL @font-face so
// visual snapshots rasterize from the same outlines on every host.
func (s *TemplatesSuite) injectEmbeddedFont(repoRoot, srcdoc string) string {
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

// runTemplateSnapshot renders the template's srcdoc in a clean tab with an embedded font
// and asserts it against the committed baseline. readySelector identifies an element that
// is only present once react-email has finished populating the preview iframe — waiting on
// it prevents capturing a partial or empty srcdoc.
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
