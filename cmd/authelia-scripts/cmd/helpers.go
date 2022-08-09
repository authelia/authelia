package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/utils"
)

type build struct {
	Branch string
	Tag    string
	Commit string
	Tagged bool
	Clean  bool
	Extra  string
	Number int
	Date   time.Time
}

// States returns the state tags for this build.
func (b build) States() []string {
	var states []string

	if b.Tagged {
		states = append(states, "tagged")
	} else {
		states = append(states, "untagged")
	}

	if b.Clean {
		states = append(states, "clean")
	} else {
		states = append(states, "dirty")
	}

	return states
}

// State returns the state tags string for this build.
func (b build) State() string {
	return strings.Join(b.States(), " ")
}

// XFlags returns the XFlags.
func (b build) XFlags() []string {
	return []string{
		fmt.Sprintf(fmtLDFLAGSX, "BuildBranch", b.Branch),
		fmt.Sprintf(fmtLDFLAGSX, "BuildTag", b.Tag),
		fmt.Sprintf(fmtLDFLAGSX, "BuildCommit", b.Commit),
		fmt.Sprintf(fmtLDFLAGSX, "BuildDate", b.Date.Format("Mon, 02 Jan 2006 15:04:05 -0700")),
		fmt.Sprintf(fmtLDFLAGSX, "BuildState", b.State()),
		fmt.Sprintf(fmtLDFLAGSX, "BuildExtra", b.Extra),
		fmt.Sprintf(fmtLDFLAGSX, "BuildNumber", strconv.Itoa(b.Number)),
	}
}

// ContainerLabels returns the container labels.
func (b build) ContainerLabels() (labels map[string]string) {
	var version string

	switch {
	case b.Clean && b.Tagged:
		version = utils.VersionAdv(b.Tag, b.State(), b.Commit, b.Branch, b.Extra)
	case b.Clean:
		version = fmt.Sprintf("%s-pre+%s.%s", b.Tag, b.Branch, b.Commit)
	case b.Tagged:
		version = fmt.Sprintf("%s-dirty", b.Tag)
	default:
		version = fmt.Sprintf("%s-dirty+%s.%s", b.Tag, b.Branch, b.Commit)
	}

	if strings.HasPrefix(version, "v") {
		version = version[:1]
	}

	labels = map[string]string{
		"org.opencontainers.image.created":       b.Date.Format(time.RFC3339),
		"org.opencontainers.image.authors":       "",
		"org.opencontainers.image.url":           "https://github.com/authelia/authelia/pkgs/container/authelia",
		"org.opencontainers.image.documentation": "https://www.authelia.com",
		"org.opencontainers.image.source":        fmt.Sprintf("https://github.com/authelia/authelia/tree/%s", b.Commit),
		"org.opencontainers.image.version":       version,
		"org.opencontainers.image.revision":      b.Commit,
		"org.opencontainers.image.vendor":        "Authelia",
		"org.opencontainers.image.licenses":      "Apache-2.0",
		"org.opencontainers.image.ref.name":      "",
		"org.opencontainers.image.title":         "authelia",
		"org.opencontainers.image.description":   "Authelia is an open-source authentication and authorization server providing two-factor authentication and single sign-on (SSO) for your applications via a web portal.",
		"org.opencontainers.image.base.digest":   "",
		"org.opencontainers.image.base.name":     "",
	}

	return labels
}

func getBuild(branch, buildNumber, extra string) (b *build, err error) {
	var out string

	b = &build{
		Branch: branch,
		Extra:  extra,
	}

	if buildNumber != "" {
		if b.Number, err = strconv.Atoi(buildNumber); err != nil {
			return nil, fmt.Errorf("error parsing provided build number: %w", err)
		}
	}

	if b.Branch == "" {
		if out, _, err = utils.RunCommandAndReturnOutput("git rev-parse --abbrev-ref HEAD"); err != nil {
			return nil, fmt.Errorf("error getting branch with git rev-parse: %w", err)
		}

		if out == "" {
			b.Branch = "master"
		} else {
			b.Branch = out
		}
	}

	var (
		gitTagCommit string
	)

	if gitTagCommit, _, err = utils.RunCommandAndReturnOutput("git rev-list --tags --max-count=1"); err != nil {
		return nil, fmt.Errorf("error getting tag commit with git rev-list: %w", err)
	}

	if b.Tag, _, err = utils.RunCommandAndReturnOutput(fmt.Sprintf("git describe --tags --abbrev=0 %s", gitTagCommit)); err != nil {
		return nil, fmt.Errorf("error getting tag with git describe: %w", err)
	}

	if b.Commit, _, err = utils.RunCommandAndReturnOutput("git rev-parse HEAD"); err != nil {
		return nil, fmt.Errorf("error getting commit with git rev-parse: %w", err)
	}

	if gitTagCommit == b.Commit {
		b.Tagged = true
	}

	if _, exitCode, _ := utils.RunCommandAndReturnOutput("git diff --quiet"); exitCode == 0 {
		b.Clean = true
	}

	b.Date = time.Now()

	return b, nil
}
