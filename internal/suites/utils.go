package suites

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-rod/rod"
	"github.com/google/uuid"
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
