package suites

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

var browserPaths = []string{"/usr/bin/chromium-browser", "/usr/bin/chromium"}

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

	if info, err = os.Stat(path); err != nil {
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
	} else {
		prefix += "/"
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

func (s *BaseSuite) SetupSuite() {
	s.SetupLogging()
	s.SetupEnvironment()
}

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

func (rs *RodSession) collectScreenshot(err error, page *rod.Page) {
	if err == context.DeadlineExceeded && os.Getenv("CI") == t {
		base := "/buildkite/screenshots"
		build := os.Getenv("BUILDKITE_BUILD_NUMBER")
		suite := strings.ToLower(os.Getenv("SUITE"))
		job := os.Getenv("BUILDKITE_JOB_ID")
		path := filepath.Join(base, build, suite, job)

		if err := os.MkdirAll(path, 0755); err != nil {
			log.Fatal(err)
		}

		pc, _, _, _ := runtime.Caller(2)
		fn := runtime.FuncForPC(pc)
		p := "github.com/authelia/authelia/v4/internal/suites."
		r := strings.NewReplacer(p, "", "(", "", ")", "", "*", "", ".", "-")

		page.MustScreenshotFullPage(fmt.Sprintf("%s/%s.jpg", path, r.Replace(fn.Name())))
	}
}

func (s *RodSuite) GetCookieNames() (names []string) {
	cookies, err := s.Page.Cookies(nil)
	s.Require().NoError(err)

	for _, cookie := range cookies {
		names = append(names, cookie.Name)
	}

	return names
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
		read, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		wd, _ := os.Getwd()
		ciPath := strings.TrimSuffix(wd, "internal/suites")
		content := strings.ReplaceAll(string(read), "/node/src/app/", ciPath+"web/")

		err = os.WriteFile(path, []byte(content), 0)
		if err != nil {
			return err
		}
	}

	return nil
}

// getEnvInfoFromURL gets environments variables for specified cookie domain
// this func makes a http call to https://login.<domain>/devworkflow and is only useful for suite tests.
func getDomainEnvInfo(domain string) (info map[string]string, err error) {
	info = make(map[string]string)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec
			},
		},
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

// generateDevEnvFile generates web/.env.development based on opts.
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

// updateDevEnvFileForDomain updates web/.env.development.
// this function only affects local dev environments.
func updateDevEnvFileForDomain(domain string, setup bool) (err error) {
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

	if err = generateDevEnvFile(info); err != nil {
		return err
	}

	if !setup {
		if err = waitUntilAutheliaFrontendIsReady(multiCookieDomainDockerEnvironment); err != nil {
			return err
		}
	}

	return nil
}
