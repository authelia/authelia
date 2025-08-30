package utils

import (
	"crypto/x509"

	"golang.org/x/text/language"
)

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
	Display    string       `json:"display"`
	Locale     string       `json:"locale"`
	Namespaces []string     `json:"namespaces,omitempty"`
	Fallbacks  []string     `json:"fallbacks,omitempty"`
	Parent     string       `json:"parent"`
	Tag        language.Tag `json:"-"`
}

type X509SystemCertPoolFactory interface {
	SystemCertPool() (pool *x509.CertPool, err error)
}

type StandardX509SystemCertPoolFactory struct{}

func (StandardX509SystemCertPoolFactory) SystemCertPool() (*x509.CertPool, error) {
	return x509.SystemCertPool()
}
