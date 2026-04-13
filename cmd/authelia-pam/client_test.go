package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

const (
	testSessionID  = "test-session-id"
	testCookieName = "authelia_session"
	testStatusOK   = `{"status":"OK"}`
	testStatusKO   = `{"status":"KO"}`
)

func newTestServer(handler http.Handler) (*httptest.Server, *Config) {
	srv := httptest.NewTLSServer(handler)

	u, _ := url.Parse(srv.URL)
	cfg := &Config{
		URL:        u,
		AuthLevel:  AuthLevel1FA2FA,
		CookieName: testCookieName,
		Timeout:    10 * time.Second,
	}

	return srv, cfg
}

func newTestClient(srv *httptest.Server, cfg *Config) *AutheliaClient {
	return &AutheliaClient{
		client:     srv.Client(),
		baseURL:    srv.URL,
		cookieName: cfg.CookieName,
	}
}

func TestFirstFactorSuccess(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/firstfactor", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if body["username"] != "testuser" || body["password"] != "testpass" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"status":"KO"}`)

			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:  testCookieName,
			Value: testSessionID,
		})

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"OK"}`)
	})

	srv, cfg := newTestServer(mux)
	defer srv.Close()

	client := newTestClient(srv, cfg)

	if err := client.FirstFactor("testuser", "testpass"); err != nil {
		t.Fatalf("FirstFactor() unexpected error: %v", err)
	}

	if client.sessionCookie != testSessionID {
		t.Errorf("sessionCookie = %q, want %q", client.sessionCookie, testSessionID)
	}
}

func TestFirstFactorFailure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/firstfactor", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"status":"KO","message":"Authentication failed."}`)
	})

	srv, cfg := newTestServer(mux)
	defer srv.Close()

	client := newTestClient(srv, cfg)

	if err := client.FirstFactor("baduser", "badpass"); err == nil {
		t.Fatal("FirstFactor() expected error, got nil")
	}
}

func TestSecondFactorTOTPSuccess(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/secondfactor/totp", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(testCookieName)
		if err != nil || cookie.Value != testSessionID {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"status":"KO"}`)

			return
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if body["token"] != "123456" {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"status":"KO"}`)

			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"OK"}`)
	})

	srv, cfg := newTestServer(mux)
	defer srv.Close()

	client := newTestClient(srv, cfg)
	client.sessionCookie = testSessionID

	if err := client.SecondFactorTOTP("123456"); err != nil {
		t.Fatalf("SecondFactorTOTP() unexpected error: %v", err)
	}
}

func TestSecondFactorTOTPInvalidToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/secondfactor/totp", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"status":"KO","message":"Authentication failed."}`)
	})

	srv, cfg := newTestServer(mux)
	defer srv.Close()

	client := newTestClient(srv, cfg)
	client.sessionCookie = testSessionID

	if err := client.SecondFactorTOTP("000000"); err == nil {
		t.Fatal("SecondFactorTOTP() expected error, got nil")
	}
}

func TestUserInfo(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/user/info", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(testCookieName)
		if err != nil || cookie.Value != testSessionID {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"OK","data":{"display_name":"Test User","method":"totp","has_totp":true,"has_webauthn":false,"has_duo":false}}`)
	})

	srv, cfg := newTestServer(mux)
	defer srv.Close()

	client := newTestClient(srv, cfg)
	client.sessionCookie = testSessionID

	info, err := client.UserInfo()
	if err != nil {
		t.Fatalf("UserInfo() unexpected error: %v", err)
	}

	if info.Method != "totp" {
		t.Errorf("Method = %q, want %q", info.Method, "totp")
	}

	if !info.HasTOTP {
		t.Error("HasTOTP = false, want true")
	}

	if info.HasWebAuthn {
		t.Error("HasWebAuthn = true, want false")
	}
}

func TestSecondFactorDuoPush(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/secondfactor/duo", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(testCookieName)
		if err != nil || cookie.Value != testSessionID {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"OK"}`)
	})

	srv, cfg := newTestServer(mux)
	defer srv.Close()

	client := newTestClient(srv, cfg)
	client.sessionCookie = testSessionID

	if err := client.SecondFactorDuoPush(); err != nil {
		t.Fatalf("SecondFactorDuoPush() unexpected error: %v", err)
	}
}

func TestRateLimiting(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/secondfactor/totp", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})

	srv, cfg := newTestServer(mux)
	defer srv.Close()

	client := newTestClient(srv, cfg)
	client.sessionCookie = testSessionID

	err := client.SecondFactorTOTP("123456")
	if err == nil {
		t.Fatal("expected rate limit error, got nil")
	}

	if err.Error() != "rate limited by Authelia server, try again later" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSpecialCharsInPassword(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/firstfactor", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if body["password"] != `p@ss"word\with/special` {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"status":"KO","received":"%s"}`, body["password"])

			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:  testCookieName,
			Value: "session-123",
		})

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"OK"}`)
	})

	srv, cfg := newTestServer(mux)
	defer srv.Close()

	client := newTestClient(srv, cfg)

	if err := client.FirstFactor("user", `p@ss"word\with/special`); err != nil {
		t.Fatalf("FirstFactor() with special chars: %v", err)
	}
}
