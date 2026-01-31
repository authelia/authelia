package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/utils"
)

// HostEntry represents an entry in /etc/hosts.
type HostEntry struct {
	Domain string
	IP     string
}

// DockerImages represents some of the data from the docker images API.
type DockerImages []DockerImage

// DockerImage represents some of the data from the docker images API.
type DockerImage struct {
	Architecture string `json:"architecture"`
	Variant      any    `json:"variant"`
	Digest       string `json:"digest"`
	OS           string `json:"os"`
}

// Match returns true if this image matches the platform.
func (d DockerImage) Match(platform string) bool {
	parts := []string{d.OS, d.Architecture} //nolint:prealloc

	if strings.Join(parts, "/") == platform {
		return true
	}

	if d.Variant == nil {
		return false
	}

	parts = append(parts, d.Variant.(string))

	return strings.Join(parts, "/") == platform
}

// Build represents a builds metadata.
type Build struct {
	Branch string
	Tag    string
	Commit string
	Tagged bool
	Clean  bool
	Extra  string
	Number int
	Date   time.Time
}

// States returns the state tags for this Build.
func (b Build) States() []string {
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

// State returns the state tags string for this Build.
func (b Build) State() string {
	return strings.Join(b.States(), " ")
}

// XFlags returns the XFlags for this Build.
func (b Build) XFlags() []string {
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

// ContainerLabels returns the container labels for this Build.
func (b Build) ContainerLabels() (labels map[string]string) {
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

	if strings.HasPrefix(version, "v") && len(version) > 1 {
		version = version[1:]
	}

	labels = map[string]string{
		"org.opencontainers.image.created":       b.Date.Format(time.RFC3339),
		"org.opencontainers.image.authors":       "Authelia Team <team@authelia.com>",
		"org.opencontainers.image.url":           "https://github.com/authelia/authelia/pkgs/container/authelia",
		"org.opencontainers.image.documentation": "https://www.authelia.com",
		"org.opencontainers.image.source":        fmt.Sprintf("https://github.com/authelia/authelia/tree/%s", b.Commit),
		"org.opencontainers.image.version":       version,
		"org.opencontainers.image.revision":      b.Commit,
		"org.opencontainers.image.vendor":        "Authelia",
		"org.opencontainers.image.licenses":      "Apache-2.0",
		"org.opencontainers.image.ref.name":      version,
		"org.opencontainers.image.title":         "authelia",
		"org.opencontainers.image.description":   "Authelia is an open-source authentication and authorization server and portal fulfilling the identity and access management (IAM) role of information security in providing multi-factor authentication and single sign-on (SSO) for your applications via a web portal. Authelia is an OpenID Connect 1.0 Provider which is OpenID Certified™ allowing comprehensive integrations and acts as a companion for common reverse proxies.",
	}

	return labels
}
