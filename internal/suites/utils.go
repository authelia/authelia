package suites

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/go-rod/rod"
	"github.com/google/uuid"
)

// GetLoginBaseURL returns the URL of the login portal and the path prefix if specified.
func GetLoginBaseURL(baseDomain string) string {
	if PathPrefix != "" {
		return LoginBaseURLFmt(baseDomain) + PathPrefix
	}

	return LoginBaseURLFmt(baseDomain)
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

// getEnvInfoFromURL gets environments variables for specified cookie domain
// this func makes a http call to https://login.<domain>/override and is only useful for suite tests.
func getDomainEnvInfo(domain string) (map[string]string, error) {
	info := make(map[string]string)

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
		err  error
	)

	targetURL := LoginBaseURLFmt(domain) + "/override"

	if req, err = http.NewRequest(http.MethodGet, targetURL, nil); err != nil {
		return info, err
	}

	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", domain)

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
func generateDevEnvFile(opts map[string]string) error {
	wd, _ := os.Getwd()
	path := strings.TrimSuffix(wd, "internal/suites")

	src := fmt.Sprintf("%s/web/.env.production", path)
	dst := fmt.Sprintf("%s/web/.env.development", path)

	tmpl, err := template.ParseFiles(src)
	if err != nil {
		return err
	}

	file, _ := os.Create(dst)
	defer file.Close()

	if err := tmpl.Execute(file, opts); err != nil {
		return err
	}

	return nil
}

// updateDevEnvFileForDomain updates web/.env.development.
func updateDevEnvFileForDomain(domain string) error {
	if os.Getenv("CI") == "true" {
		return nil
	}

	info, err := getDomainEnvInfo(domain)
	if err != nil {
		return err
	}

	err = generateDevEnvFile(info)
	if err != nil {
		return err
	}

	return nil
}
