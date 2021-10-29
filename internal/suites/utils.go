package suites

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/matryer/is"

	"github.com/authelia/authelia/v4/internal/storage"
)

// GetLoginBaseURL returns the URL of the login portal and the path prefix if specified.
func GetLoginBaseURL() string {
	if PathPrefix != "" {
		return LoginBaseURL + PathPrefix
	}

	return LoginBaseURL
}

func (rs *RodSession) collectCoverage(page *rod.Page) {
	coverageDir := "../../web/.nyc_output"
	now := time.Now()

	resp, err := page.Eval("() => JSON.stringify(window.__coverage__)")
	if err != nil {
		log.Fatal(err)
	}

	coverageData := fmt.Sprintf("%v", resp.Value)

	_ = os.MkdirAll(coverageDir, 0775)

	if coverageData != "<nil>" {
		err = os.WriteFile(fmt.Sprintf("%s/coverage-%d.json", coverageDir, now.Unix()), []byte(coverageData), 0664) //nolint:gosec
		if err != nil {
			log.Fatal(err)
		}

		err = filepath.Walk("../../web/.nyc_output", fixCoveragePath)
		if err != nil {
			log.Fatal(err)
		}
	}
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

func setupTest(t *testing.T, proxy string, register bool) RodSuite {
	s := RodSuite{}

	browser, err := StartRodWithProxy(proxy)
	if err != nil {
		log.Fatal(err)
	}

	s.RodSession = browser

	if proxy == "" {
		s.Page = s.doCreateTab(HomeBaseURL)
		s.verifyIsHome(t, s.Page)
	}

	if register {
		secret = s.doLoginAndRegisterTOTP(t, s.Page, testUsername, testPassword, false)
	}

	return s
}

func setupCLITest() (s *CommandSuite) {
	s = &CommandSuite{}

	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/docker-compose.yml",
		"internal/suites/CLI/docker-compose.yml",
		"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
	})
	s.DockerEnvironment = dockerEnvironment

	testArg := ""
	coverageArg := ""

	if os.Getenv("CI") == t {
		testArg = "-test.coverprofile=/authelia/coverage-$(date +%s).txt"
		coverageArg = "COVERAGE"
	}

	s.testArg = testArg
	s.coverageArg = coverageArg

	return s
}

func teardownTest(s RodSuite) {
	s.collectCoverage(s.Page)
	s.MustClose()
	err := s.RodSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func teardownDuoTest(t *testing.T, s RodSuite) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.collectScreenshot(ctx.Err(), s.Page)
	}()

	is := is.New(t)
	provider := storage.NewSQLiteProvider(&storageLocalTmpConfig)
	is.NoErr(provider.SavePreferred2FAMethod(ctx, "john", "totp"))
	is.NoErr(provider.DeletePreferredDuoDevice(ctx, "john"))
}
