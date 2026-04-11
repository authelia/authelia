//go:build externalsuites
// +build externalsuites

package suites

import (
	"context"
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
		globalDevServer = early
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

	globalDevServer = nil
}

func (s *TemplatesSuite) templatesURL(path string) string {
	return s.baseURL + path
}

// openPreviewFrame loads /preview/<component> and returns the rod *Page representing the
// same-origin srcdoc iframe the preview server renders the email into.
func (s *TemplatesSuite) openPreviewFrame(outer *rod.Page) *rod.Page {
	iframeEl := s.WaitElementLocatedBySelector(s.T(), outer, "iframe")

	frame, err := iframeEl.Frame()
	require.NoError(s.T(), err, "failed to descend into preview iframe")

	_, err = frame.Element("body")
	require.NoError(s.T(), err, "preview iframe has no body yet")

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

	frame := s.openPreviewFrame(outer)

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

	frame := s.openPreviewFrame(outer)

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

	frame := s.openPreviewFrame(outer)

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

// runTemplateSnapshot renders the given template's pristine srcdoc markup in a clean tab at
// a fixed viewport, bypassing the preview server's resize / dark-mode chrome so the captured
// image is deterministic run-to-run, and asserts it against the committed baseline.
func (s *TemplatesSuite) runTemplateSnapshot(slug, snapshotName string) {
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

	iframeEl := s.WaitElementLocatedBySelector(s.T(), outer, "iframe")

	srcdocAttr := iframeEl.MustAttribute("srcdoc")
	require.NotNil(s.T(), srcdocAttr, "expected preview iframe to have a srcdoc attribute")

	clean.MustSetViewport(800, 1200, 1, false)

	require.NoError(s.T(), clean.SetDocumentContent(*srcdocAttr))

	s.WaitForVisualStable(s.T(), clean)

	screenshot := s.FullPageScreenshot(s.T(), clean)

	repoRoot, err := findRepoRoot()
	require.NoError(s.T(), err)

	AssertVisualSnapshot(s.T(), repoRoot, snapshotName, screenshot)
}

func (s *TemplatesSuite) TestIdentityVerificationOTCVisualSnapshot() {
	s.runTemplateSnapshot("IdentityVerificationOTC", "templates_identity_verification_otc_snapshot.png")
}

func (s *TemplatesSuite) TestIdentityVerificationJWTVisualSnapshot() {
	s.runTemplateSnapshot("IdentityVerificationJWT", "templates_identity_verification_jwt_snapshot.png")
}

func (s *TemplatesSuite) TestEventVisualSnapshot() {
	s.runTemplateSnapshot("Event", "templates_event_snapshot.png")
}
