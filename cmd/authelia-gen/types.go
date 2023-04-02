// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/language"
)

type tmplIssueTemplateData struct {
	Labels   []string
	Versions []string
	Proxies  []string
}

type tmplConfigurationKeysData struct {
	Timestamp time.Time
	Keys      []string
	Package   string
}

type tmplScriptsGEnData struct {
	Package          string
	VersionSwaggerUI string
}

// GitHubTagsJSON represents the JSON struct for the GitHub Tags API.
type GitHubTagsJSON struct {
	Name string `json:"name"`
}

type GitHubReleasesJSON struct {
	ID              int              `json:"id"`
	Name            string           `json:"name"`
	TagName         string           `json:"tag_name"`
	TargetCommitISH string           `json:"target_commitish"`
	NodeID          string           `json:"node_id"`
	Draft           bool             `json:"draft"`
	Prerelease      bool             `json:"prerelease"`
	URL             string           `json:"url"`
	AssetsURL       string           `json:"assets_url"`
	UploadURL       string           `json:"upload_url"`
	HTMLURL         string           `json:"html_url"`
	TarballURL      string           `json:"tarball_url"`
	ZipballURL      string           `json:"zipball_url"`
	Assets          []any            `json:"assets"`
	CreatedAt       time.Time        `json:"created_at"`
	PublishedAt     time.Time        `json:"published_at"`
	Author          GitHubAuthorJSON `json:"author"`
	Body            string           `json:"body"`
}

type GitHubAuthorJSON struct {
	ID                int    `json:"id"`
	Login             string `json:"login"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

// DocsDataMisc represents the docs misc data schema.
type DocsDataMisc struct {
	CSP    TemplateCSP `json:"csp"`
	Latest string      `json:"latest"`
}

// TemplateCSP represents the CSP template vars.
type TemplateCSP struct {
	TemplateDefault     string `json:"default"`
	TemplateDevelopment string `json:"development"`
	PlaceholderNONCE    string `json:"nonce"`
}

// ConfigurationKey is the docs json model for the Authelia configuration keys.
type ConfigurationKey struct {
	Path   string `json:"path"`
	Secret bool   `json:"secret"`
	Env    string `json:"env"`
}

// Languages is the docs json model for the Authelia languages configuration.
type Languages struct {
	Defaults   DefaultsLanguages `json:"defaults"`
	Namespaces []string          `json:"namespaces"`
	Languages  []Language        `json:"languages"`
}

type DefaultsLanguages struct {
	Language  Language `json:"language"`
	Namespace string   `json:"namespace"`
}

// Language is the docs json model for a language.
type Language struct {
	Display    string   `json:"display"`
	Locale     string   `json:"locale"`
	Namespaces []string `json:"namespaces,omitempty"`
	Fallbacks  []string `json:"fallbacks,omitempty"`

	Tag language.Tag `json:"-"`
}

const (
	labelAreaPrefixPriority = "priority"
	labelAreaPrefixType     = "type"
	labelAreaPrefixStatus   = "status"
)

type labelPriority int

//nolint:deadcode,varcheck // Kept for future use.
const (
	labelPriorityCritical labelPriority = iota
	labelPriorityHigh
	labelPriorityMedium
	labelPriorityNormal
	labelPriorityLow
)

var labelPriorityDescriptions = [...]string{
	"Critical",
	"High",
	"Medium",
	"Normal",
	"Low",
}

func (p labelPriority) String() string {
	return fmt.Sprintf("%s/%d/%s", labelAreaPrefixPriority, p+1, strings.ToLower(labelPriorityDescriptions[p]))
}

func (p labelPriority) Description() string {
	return labelPriorityDescriptions[p]
}

type labelStatus int

const (
	labelStatusNeedsDesign labelStatus = iota
	labelStatusNeedsTriage
)

var labelStatusDescriptions = [...]string{
	"needs-design",
	"needs-triage",
}

func (s labelStatus) String() string {
	return fmt.Sprintf("%s/%s", labelAreaPrefixStatus, labelStatusDescriptions[s])
}

type labelType int

//nolint:deadcode,varcheck // Kept for future use.
const (
	labelTypeFeature labelType = iota
	labelTypeBugUnconfirmed
	labelTypeBug
)

var labelTypeDescriptions = [...]string{
	"feature",
	"bug/unconfirmed",
	"bug",
}

func (t labelType) String() string {
	return fmt.Sprintf("%s/%s", labelAreaPrefixType, labelTypeDescriptions[t])
}

// CSPValue represents individual CSP values.
type CSPValue struct {
	Name  string
	Value string
}

// PackageJSON represents a NPM package.json file.
type PackageJSON struct {
	Version string `json:"version"`
}
