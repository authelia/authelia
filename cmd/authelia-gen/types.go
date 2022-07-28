package main

import (
	"fmt"
	"time"
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

// Languages is the docs json model for the Authelia languages configuration.
type Languages struct {
	DefaultLocale    string     `json:"defaultLocale"`
	DefaultNamespace string     `json:"defaultNamespace"`
	Namespaces       []string   `json:"namespaces"`
	Languages        []Language `json:"languages"`
}

// Language is the docs json model for a language.
type Language struct {
	Display    string   `json:"display"`
	Locale     string   `json:"locale"`
	Namespaces []string `json:"namespaces"`
	Fallbacks  []string `json:"fallbacks"`
}

const (
	labelAreaPrefixPriority = "priority"
	labelAreaPrefixType     = "type"
	labelAreaPrefixStatus   = "status"
)

type labelPriority int

const (
	labelPriorityCritical labelPriority = iota
	labelPriorityHigh
	labelPriorityMedium
	labelPriorityNormal
	labelPriorityLow
)

var labelPriorityDescriptions = [...]string{
	"critical",
	"high",
	"medium",
	"normal",
	"low",
}

func (p labelPriority) String() string {
	return fmt.Sprintf("%s/%d-%s", labelAreaPrefixPriority, p+1, labelPriorityDescriptions[p])
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

const (
	labelTypeFeature labelType = iota
	labelTypePotentialBug
	labelTypeBug
)

var labelTypeDescriptions = [...]string{
	"feature",
	"potential-bug",
	"bug",
}

func (t labelType) String() string {
	return fmt.Sprintf("%s/%s", labelAreaPrefixType, labelTypeDescriptions[t])
}
