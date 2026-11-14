// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package appauthenticator

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	androidStatusFresh   = 24 * time.Hour
	androidStatusMaxAge  = 7 * 24 * time.Hour
	androidStatusMaxSize = 32 * 1024 * 1024
	androidStatusBackoff = time.Minute
	androidStatusTimeout = 10 * time.Second
)

type androidRevocation struct {
	url    string
	client *http.Client

	mu      sync.Mutex
	entries map[string]string
	fetched time.Time
	failed  time.Time
}

type androidStatusList struct {
	Entries map[string]struct {
		Status string `json:"status"`
	} `json:"entries"`
}

func checkAndroidRevocation(statuses func() (map[string]string, error), certs []*x509.Certificate) error {
	entries, err := statuses()
	if err != nil {
		return fmt.Errorf("error checking revocation: %w", err)
	}

	for i, cert := range certs {
		for _, key := range []string{cert.SerialNumber.Text(16), cert.SerialNumber.Text(10)} {
			if status, ok := entries[key]; ok {
				return fmt.Errorf("certificate %d has revocation status '%s'", i, status)
			}
		}
	}

	return nil
}

func (r *androidRevocation) list(ctx context.Context, now time.Time) (map[string]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.entries != nil && now.Sub(r.fetched) < androidStatusFresh {
		return r.entries, nil
	}

	usable := r.entries != nil && now.Sub(r.fetched) <= androidStatusMaxAge

	if usable && !r.failed.IsZero() && now.Sub(r.failed) < androidStatusBackoff {
		return r.entries, nil
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), androidStatusTimeout)
	defer cancel()

	entries, err := r.fetch(ctx)
	if err == nil {
		r.entries, r.fetched, r.failed = entries, now, time.Time{}

		return entries, nil
	}

	r.failed = now

	if usable {
		return r.entries, nil
	}

	return nil, fmt.Errorf("no usable attestation status list: %w", err)
}

func (r *androidRevocation) fetch(ctx context.Context) (map[string]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching status list: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var list androidStatusList

	if err = json.NewDecoder(io.LimitReader(resp.Body, androidStatusMaxSize)).Decode(&list); err != nil {
		return nil, fmt.Errorf("error decoding status list: %w", err)
	}

	if list.Entries == nil {
		return nil, errors.New("status list has no entries")
	}

	entries := make(map[string]string, len(list.Entries))

	for key, entry := range list.Entries {
		entries[strings.ToLower(key)] = entry.Status
	}

	return entries, nil
}
