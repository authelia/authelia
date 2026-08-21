package suites

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/utils"
)

var browserPaths = []string{"/usr/bin/chromium-browser", "/usr/bin/chromium", "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", "/Applications/Chromium.app/Contents/MacOS/Chromium"}

const screenshotBanner = `() => {
	const height = 28;
	const banner = document.createElement('div');

	banner.textContent = location.href;
	banner.style.cssText = 'position:fixed;top:0;left:0;right:0;z-index:2147483647;' +
		'height:' + height + 'px;box-sizing:border-box;' +
		'background:#1f2937;color:#f9fafb;font:12px/20px ui-monospace,monospace;' +
		'padding:4px 8px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;' +
		'border-bottom:1px solid #4b5563;';

	document.documentElement.style.paddingTop = height + 'px';
	document.documentElement.appendChild(banner);
}`

const diagnosticsResources = `() => JSON.stringify({
	url: location.href,
	title: document.title,
	readyState: document.readyState,
	bodyLength: document.body ? document.body.innerHTML.length : 0,
	resources: performance.getEntriesByType('resource').map((entry) => ({
		name: entry.name,
		initiator: entry.initiatorType,
		status: entry.responseStatus,
		transferSize: entry.transferSize,
		encodedBodySize: entry.encodedBodySize,
		duration: Math.round(entry.duration),
	})),
}, null, 2)`

// StringToKeys returns the input.Key values which represent the given string.
func StringToKeys(value string) []input.Key {
	n := len(value)

	keys := make([]input.Key, n)

	for i := 0; i < n; i++ {
		keys[i] = input.Key(value[i])
	}

	return keys
}

// ValidateBrowserPath validates the appropriate chromium browser path.
func ValidateBrowserPath(path string) (browserPath string, err error) {
	var info os.FileInfo

	if info, err = os.Stat(path); err != nil { //nolint:gosec // TODO: Run this line through taint analysis.
		return "", err
	} else if info.IsDir() {
		return "", fmt.Errorf("browser cannot be a directory")
	}

	return path, nil
}

// GetBrowserPath retrieves the appropriate chromium browser path.
func GetBrowserPath() (path string, err error) {
	browserPath := os.Getenv("BROWSER_PATH")

	if browserPath != "" {
		return ValidateBrowserPath(browserPath)
	}

	for _, browserPath = range browserPaths {
		if browserPath, err = ValidateBrowserPath(browserPath); err == nil {
			return browserPath, nil
		}
	}

	return "", fmt.Errorf("no chromium browser was detected in the known paths, set the BROWSER_PATH environment variable to override the path")
}

// GetLoginBaseURL returns the URL of the login portal and the path prefix if specified.
func GetLoginBaseURL(baseDomain string) string {
	return LoginBaseURLFmt(baseDomain) + GetPathPrefix()
}

// GetLoginBaseURLWithFallbackPrefix overloads GetLoginBaseURL and includes '/' as a prefix if the prefix is empty.
func GetLoginBaseURLWithFallbackPrefix(baseDomain, fallback string) string {
	prefix := GetPathPrefix()

	if prefix == "" {
		prefix = fallback
	}

	return LoginBaseURLFmt(baseDomain) + prefix
}

func (rs *RodSession) collectCoverage(page *rod.Page) {
	coverageDir := "../../web/.nyc_output"

	resp, err := page.Eval("() => JSON.stringify(window.__coverage__)")
	if err != nil {
		log.Fatal(err)
	}

	coverageData := fmt.Sprintf("%v", resp.Value)

	_ = os.MkdirAll(coverageDir, 0775)

	if coverageData != "<nil>" {
		err = os.WriteFile(fmt.Sprintf("%s/coverage-%s.json", coverageDir, uuid.New().String()), []byte(coverageData), 0664) //nolint:gosec
		if err != nil {
			log.Fatal(err)
		}

		err = filepath.Walk("../../web/.nyc_output", fixCoveragePath)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// SetupSuite performs the setup for this suite.
func (s *BaseSuite) SetupSuite() {
	s.SetupLogging()
	s.SetupEnvironment()
}

// SetupLogging configures the logging for this suite.
func (s *BaseSuite) SetupLogging() {
	if os.Getenv("SUITE_SETUP_LOGGING") == t {
		return
	}

	var (
		level string
		ok    bool
	)

	if level, ok = os.LookupEnv("SUITES_LOG_LEVEL"); !ok {
		return
	}

	l, err := log.ParseLevel(level)

	s.NoError(err)

	log.SetLevel(l)

	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})

	s.T().Setenv("SUITE_SETUP_LOGGING", t)
}

// SetupEnvironment configures the environment for this suite.
func (s *BaseSuite) SetupEnvironment() {
	if s.Name == "" || os.Getenv("SUITE_SETUP_ENVIRONMENT") == t {
		return
	}

	log.Debugf("Checking Suite %s for .env file", s.Name)

	path := filepath.Join(s.Name, ".env")

	var (
		info os.FileInfo
		err  error
	)

	path, err = filepath.Abs(path)

	s.Require().NoError(err)

	if info, err = os.Stat(path); err != nil {
		s.Assert().True(os.IsNotExist(err))

		log.Debugf("Suite %s does not have an .env file or it can't be read: %v", s.Name, err)

		return
	}

	s.Require().False(info.IsDir())

	log.Debugf("Suite %s does have an .env file at path: %s", s.Name, path)

	var file *os.File

	file, err = os.Open(path)

	s.Require().NoError(err)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		v := strings.Split(scanner.Text(), "=")

		s.Require().Len(v, 2)

		s.T().Setenv(v[0], v[1])
	}

	s.T().Setenv("SUITE_SETUP_ENVIRONMENT", t)
}

// SuiteTempDir creates a directory inside the directory this process shares with the suite containers, and returns the
// path this process reaches it at. Pass the result through SuiteTmpContainerPath before handing it to a container.
// t.TempDir() cannot be used because it honors TMPDIR, which the containers do not share.
func SuiteTempDir(t *testing.T, pattern string) (dir string) {
	dir, err := os.MkdirTemp(SuiteTmpPath(), pattern)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	return dir
}

// SuiteTmpContainerPath returns the path a container sees for a file this process reaches at dir, which has to be
// inside SuiteTmpPath. The containers always see that directory at /tmp while this process may reach it elsewhere, so
// a path that crosses the boundary has to be translated rather than shared verbatim.
func SuiteTmpContainerPath(dir string) string {
	relative, err := filepath.Rel(SuiteTmpPath(), dir)
	if err != nil {
		return dir
	}

	return filepath.Join(suiteTmpPathDefault, relative)
}

func screenshotPaths(name string) (path, reported string) {
	suite := strings.ToLower(os.Getenv("SUITE"))

	if os.Getenv("CI") == t {
		// Reported relative to the repository root so it matches the Buildkite artifact path, which
		// is the form an artifact:// reference has to be given.
		reported = filepath.Join("screenshots", suite, name)

		return filepath.Join("../..", reported), reported
	}

	// Scoped by compose project so concurrent runs of the same suite on one machine keep their own screenshots.
	path = filepath.Join(os.TempDir(), "authelia-suites-screenshots", composeProjectName(), suite, name)

	return path, path
}

func (s *RodSuite) collectScreenshot(err error, page *rod.Page) {
	s.RodSession.collectScreenshot(s.T(), err, page)
}

func (rs *RodSession) collectContainerLogs(test *testing.T, base string) {
	// The OnError hook prints these too, but it runs in a separate process after the test binary has
	// exited, so nothing it prints can be associated with the test that failed. The collected tail is deep
	// because one configuration rebuild on a shared daemon costs a proxy a debug line per container the
	// daemon holds, which on its own is as many lines as this used to collect in total. Only the artifact
	// grows with the depth, since the lines reported inline stay at containerLogTailLines.
	output, _, err := utils.RunCommandAndReturnOutput(
		fmt.Sprintf("docker ps --filter label=com.docker.compose.project=%s --format '{{.Names}}'", composeProjectName()),
	)
	if err != nil {
		log.Debugf("Error listing suite containers: %v", err)

		return
	}

	var builder strings.Builder

	for _, name := range strings.Fields(output) {
		logs, _, lerr := utils.RunCommandAndReturnOutput(fmt.Sprintf("docker logs --tail %d %s 2>&1", containerLogLines, name))
		if lerr != nil {
			log.Debugf("Error reading logs of container '%s': %v", name, lerr)

			continue
		}

		fmt.Fprintf(&builder, "===== %s =====\n%s\n", name, logs)

		test.Logf("Last %d log lines of '%s':\n%s", containerLogTailLines, name, tailLines(logs, containerLogTailLines))
	}

	if builder.Len() == 0 {
		return
	}

	path, _ := screenshotPaths(base + ".containers.log")

	if err = os.WriteFile(path, []byte(builder.String()), 0600); err != nil {
		log.Debugf("Error writing '%s': %v", path, err)
	}
}

func (rs *RodSession) collectProxyAccessLog(base string) {
	source := SuiteTmpPath(proxyAccessLog())

	data, err := os.ReadFile(source)
	if err != nil {
		// Suites that do not run behind a proxy have no such file.
		log.Debugf("Error reading '%s': %v", source, err)

		return
	}

	path, _ := screenshotPaths(base + ".access.log")

	// Both paths are named by this package from the name of the running test, and neither carries anything
	// a request could reach.
	if err = os.WriteFile(path, data, 0600); err != nil { //nolint:gosec
		log.Debugf("Error writing '%s': %v", path, err)
	}
}

func tailLines(value string, n int) string {
	lines := strings.Split(strings.TrimRight(value, "\n"), "\n")

	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	return strings.Join(lines, "\n")
}

func (rs *RodSession) collectDiagnostics(page *rod.Page, base string) {
	// A screenshot of a page that rendered nothing shows nothing, so these answer whether the markup
	// arrived at all and whether any of its resources failed to load.
	for name, expression := range map[string]string{
		base + ".html":           `() => document.documentElement.outerHTML`,
		base + ".resources.json": diagnosticsResources,
	} {
		path, _ := screenshotPaths(name)

		value, err := page.Eval(expression)
		if err != nil {
			log.Debugf("Error collecting '%s': %v", name, err)

			continue
		}

		if err = os.WriteFile(path, []byte(value.Value.Str()), 0600); err != nil {
			log.Debugf("Error writing '%s': %v", path, err)
		}
	}
}

func (rs *RodSession) collectScreenshot(test *testing.T, err error, page *rod.Page) {
	if !test.Failed() && !errors.Is(err, context.DeadlineExceeded) {
		return
	}

	base := strings.NewReplacer("/", "-", " ", "_").Replace(test.Name())

	defer rs.collectContainerLogs(test, base)
	defer rs.collectProxyAccessLog(base)

	path, reported := screenshotPaths(base + ".png")

	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Errorf("Error creating screenshot directory '%s': %v", filepath.Dir(path), err)

		return
	}

	rs.collectDiagnostics(page, base)

	if _, err = page.Eval(screenshotBanner); err != nil {
		log.Debugf("Error labeling the screenshot with the page URL: %v", err)
	}

	data, err := page.Screenshot(true, nil)
	if err != nil {
		log.Errorf("Error capturing screenshot for '%s': %v", test.Name(), err)

		return
	}

	if err = os.WriteFile(path, data, 0600); err != nil {
		log.Errorf("Error writing screenshot '%s': %v", path, err)

		return
	}

	var url string

	if info, ierr := page.Info(); ierr == nil {
		url = info.URL
	}

	if build, job := os.Getenv("BUILDKITE_BUILD_URL"), os.Getenv("BUILDKITE_JOB_ID"); build != "" && job != "" {
		test.Logf("Failure screenshot of '%s' at '%s': %s/waterfall?jid=%s&tab=artifacts", url, reported, build, job)
	} else {
		test.Logf("Failure screenshot of '%s' at '%s'", url, reported)
	}
}

// GetCookieNames returns the names of the cookies currently set in the browser.
func (s *RodSuite) GetCookieNames() (names []string) {
	cookies, err := s.Cookies(nil)
	s.Require().NoError(err)

	for _, cookie := range cookies {
		names = append(names, cookie.Name)
	}

	return names
}

// VerifyPageElementAttributeValueBoolean verifies the boolean attribute of the element matching the given selector.
func (s *RodSuite) VerifyPageElementAttributeValueBoolean(t *testing.T, page *rod.Page, cssSelector, name string, required, value bool) {
	s.VerifyPageElementAttributeValue(t, page, cssSelector, name, required, strconv.FormatBool(value))
}

// VerifyPageElementAttributeValue verifies the attribute of the element matching the given selector.
func (s *RodSuite) VerifyPageElementAttributeValue(t *testing.T, page *rod.Page, cssSelector, name string, required bool, value string) {
	element := s.WaitElementLocatedByID(t, page, cssSelector)

	attr, err := element.Attribute(name)
	require.NoError(t, err)

	require.True(t, !required || attr != nil)

	if attr == nil {
		return
	}

	assert.Equal(t, value, *attr)
}

func fixCoveragePath(path string, file os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if file.IsDir() {
		return nil
	}

	coverage, err := filepath.Match("*.json", file.Name())
	if err != nil {
		return err
	}

	if coverage {
		var data []byte

		if data, err = os.ReadFile(path); err != nil { //nolint:gosec
			return err
		}

		wd, _ := os.Getwd()
		ciPath := strings.TrimSuffix(wd, "internal/suites")
		content := strings.ReplaceAll(string(data), "/node/src/app/", ciPath+"web/")

		if err = os.WriteFile(path, []byte(content), 0); err != nil { //nolint:gosec
			return err
		}
	}

	return nil
}

func getDomainEnvInfo(domain string) (info map[string]string, err error) {
	info = make(map[string]string)

	client := &http.Client{
		Transport: NewHTTPTransport(),
	}

	var (
		req  *http.Request
		resp *http.Response
		body []byte
	)

	targetURL := LoginBaseURLFmt(domain) + "/devworkflow"

	if req, err = http.NewRequest(http.MethodGet, targetURL, nil); err != nil {
		return info, err
	}

	req.Header.Set(fasthttp.HeaderXForwardedProto, "https")
	req.Header.Set(fasthttp.HeaderXForwardedHost, domain)

	if resp, err = client.Do(req); err != nil {
		return info, err
	}

	if body, err = io.ReadAll(resp.Body); err != nil {
		return info, err
	}
	defer resp.Body.Close()

	if err = json.Unmarshal(body, &info); err != nil {
		return info, err
	}

	return info, nil
}

func generateDevEnvFile(info map[string]string) (err error) {
	base, _ := os.Getwd()
	base = strings.TrimSuffix(base, "/internal/suites")

	var tmpl *template.Template

	if tmpl, err = template.ParseFiles(base + envFileProd); err != nil {
		return err
	}

	file, _ := os.Create(base + envFileDev)
	defer file.Close()

	if err = tmpl.Execute(file, info); err != nil {
		return err
	}

	return nil
}

func updateDevEnvFileForDomain(domain string, dockerEnvironment *DockerEnvironment) (err error) {
	if os.Getenv("CI") == t {
		return nil
	}

	if _, err = os.Stat(envFileDev); err != nil && os.IsNotExist(err) {
		file, _ := os.Create(envFileDev)
		file.Close()
	}

	var info map[string]string

	if info, err = getDomainEnvInfo(domain); err != nil {
		return err
	}

	since := time.Now()

	if err = generateDevEnvFile(info); err != nil {
		return err
	}

	return waitUntilAutheliaFrontendRestarted(dockerEnvironment, since)
}
