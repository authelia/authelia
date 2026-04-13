package main

import (
	"testing"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantErr   bool
		wantLevel AuthLevel
		wantURL   string
	}{
		{
			name:      "valid full config",
			args:      []string{"--url", "https://auth.example.com", "--auth-level", "1FA+2FA", "--cookie-name", "my_session", "--timeout", "60"},
			wantErr:   false,
			wantLevel: AuthLevel1FA2FA,
			wantURL:   "https://auth.example.com",
		},
		{
			name:      "valid 1fa only",
			args:      []string{"--url", "https://auth.example.com", "--auth-level", "1FA"},
			wantErr:   false,
			wantLevel: AuthLevel1FA,
		},
		{
			name:      "valid 2fa only",
			args:      []string{"--url", "https://auth.example.com", "--auth-level", "2FA"},
			wantErr:   false,
			wantLevel: AuthLevel2FA,
		},
		{
			name:      "defaults",
			args:      []string{"--url", "https://auth.example.com"},
			wantErr:   false,
			wantLevel: AuthLevel1FA2FA,
		},
		{
			name:    "missing url",
			args:    []string{"--auth-level", "1FA"},
			wantErr: true,
		},
		{
			name:    "http not allowed",
			args:    []string{"--url", "http://auth.example.com"},
			wantErr: true,
		},
		{
			name:    "invalid auth level",
			args:    []string{"--url", "https://auth.example.com", "--auth-level", "3FA"},
			wantErr: true,
		},
		{
			name:    "invalid timeout",
			args:    []string{"--url", "https://auth.example.com", "--timeout", "0"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := ParseConfig(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if cfg.AuthLevel != tt.wantLevel {
				t.Errorf("AuthLevel = %v, want %v", cfg.AuthLevel, tt.wantLevel)
			}

			if tt.wantURL != "" && cfg.URL.String() != tt.wantURL {
				t.Errorf("URL = %v, want %v", cfg.URL.String(), tt.wantURL)
			}
		})
	}
}

func TestParseConfigDefaults(t *testing.T) {
	cfg, err := ParseConfig([]string{"--url", "https://auth.example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.CookieName != "authelia_session" {
		t.Errorf("CookieName = %q, want %q", cfg.CookieName, "authelia_session")
	}

	if cfg.Timeout.Seconds() != 30 {
		t.Errorf("Timeout = %v, want 30s", cfg.Timeout)
	}

	if cfg.CACert != "" {
		t.Errorf("CACert = %q, want empty", cfg.CACert)
	}
}
