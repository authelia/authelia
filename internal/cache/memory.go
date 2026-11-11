// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/authelia/authelia/v4/internal/session"
)

// NewMemory returns a new Memory Provider with its lookups initialized.
func NewMemory() (memory *Memory) {
	return &Memory{
		session:        map[memoryKey]*itemSession{},
		lookupPublicID: map[memoryKey]string{},
		lookupUsername: map[memoryKey][]string{},
	}
}

// Memory is a Provider which stores sessions in process memory. Every exported method acquires mu, and the items stored
// in the maps are only ever read or mutated while it is held, which makes the unexported helpers below unsafe to call
// on their own.
type Memory struct {
	mu sync.RWMutex

	session        map[memoryKey]*itemSession
	lookupPublicID map[memoryKey]string
	lookupUsername map[memoryKey][]string
}

// StartupCheck implements the Provider interface.
func (m *Memory) StartupCheck() (err error) {
	return nil
}

// SessionGet implements the Provider interface.
func (m *Memory) SessionGet(ctx context.Context, issuer, id string) (record session.Record, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, ok := m.session[m.key(issuer, id)]
	if !ok || item.destroyed || item.expired(time.Now()) {
		return nil, nil
	}

	return session.NewRecord(item.id, item.data), nil
}

// SessionGetByPublicID implements the Provider interface.
func (m *Memory) SessionGetByPublicID(ctx context.Context, issuer, pid string) (record session.Record, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id, ok := m.lookupPublicID[m.key(issuer, pid)]
	if !ok {
		return nil, nil
	}

	item, ok := m.session[m.key(issuer, id)]
	if !ok || item.destroyed || item.expired(time.Now()) {
		return nil, nil
	}

	return session.NewRecord(item.id, item.data), nil
}

// SessionGetIDsByUsername implements the Provider interface.
func (m *Memory) SessionGetIDsByUsername(ctx context.Context, issuer, username string) (ids []string, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lookup := m.lookupUsername[m.key(issuer, username)]

	now, ids := time.Now(), make([]string, 0, len(lookup))

	for _, id := range lookup {
		if item, ok := m.session[m.key(issuer, id)]; ok && !item.destroyed && !item.expired(now) {
			ids = append(ids, id)
		}
	}

	return ids, nil
}

// SessionSave implements the Provider interface.
func (m *Memory) SessionSave(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if other, ok := m.lookupPublicID[m.key(issuer, pid)]; ok && other != id {
		if _, exists := m.session[m.key(issuer, other)]; exists {
			return session.ErrSessionSuperseded
		}
	}

	key := m.key(issuer, id)

	if item, ok := m.session[key]; ok {
		if item.destroyed && !item.expired(time.Now()) {
			return session.ErrSessionSuperseded
		}

		if item.pid != pid {
			delete(m.lookupPublicID, m.key(issuer, item.pid))
		}

		if item.username != username {
			m.removeUsername(id, item.username, issuer)
		}
	}

	m.session[key] = &itemSession{
		data:     data,
		id:       id,
		pid:      pid,
		issuer:   issuer,
		username: username,
		expires:  sessionExpires(expiration),
	}

	m.lookupPublicID[m.key(issuer, pid)] = id

	m.setUsername(id, username, issuer)

	return nil
}

// SessionSaveData implements the Provider interface. The public id and username are unused as the stored session
// already records them and its lookups are derived from it rather than expiring independently.
func (m *Memory) SessionSaveData(ctx context.Context, issuer, id, _, _ string, expiration time.Duration, data []byte) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.key(issuer, id)

	item, ok := m.session[key]
	if !ok {
		return fmt.Errorf("session not found: %s:%s", issuer, id)
	}

	if item.destroyed {
		return session.ErrSessionSuperseded
	}

	item.data = data
	item.expires = sessionExpires(expiration)

	return nil
}

// SessionChangeID implements the Provider interface.
func (m *Memory) SessionChangeID(ctx context.Context, issuer, oldID, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.key(issuer, oldID)

	item, ok := m.session[key]
	if !ok || item.destroyed {
		return nil
	}

	delete(m.session, key)

	if item.pid != pid {
		delete(m.lookupPublicID, m.key(issuer, item.pid))
	}

	m.removeUsername(oldID, item.username, issuer)

	if username != "" {
		item.username = username
	}

	item.data = data
	item.id = id
	item.pid = pid
	item.expires = sessionExpires(expiration)

	m.session[m.key(issuer, id)] = item
	m.lookupPublicID[m.key(issuer, pid)] = id

	m.setUsername(id, item.username, issuer)

	return nil
}

// SessionDelete implements the Provider interface.
func (m *Memory) SessionDelete(ctx context.Context, issuer, id, pid, username string) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, ok := m.session[m.key(issuer, id)]
	if !ok {
		m.delete(id, pid, username, issuer)

		return nil
	}

	if item.destroyed {
		return nil
	}

	m.delete(id, item.pid, item.username, issuer)

	item.data, item.username, item.destroyed = nil, "", true

	m.session[m.key(issuer, id)] = item

	return nil
}

// SessionGarbageCollectionFrequency implements the Provider interface. Memory does not expire records itself so a
// non zero frequency is always returned.
func (m *Memory) SessionGarbageCollectionFrequency(ctx context.Context) (frequency time.Duration) {
	return sessionGarbageCollectionFrequency
}

// SessionGarbageCollection implements the Provider interface, removing every expired session and its lookups.
func (m *Memory) SessionGarbageCollection(ctx context.Context) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	for _, item := range m.session {
		if item.expired(now) {
			m.delete(item.id, item.pid, item.username, item.issuer)
		}
	}

	return nil
}

func (m *Memory) delete(id, pid, username, issuer string) {
	delete(m.session, m.key(issuer, id))

	if key := m.key(issuer, pid); m.lookupPublicID[key] == id {
		delete(m.lookupPublicID, key)
	}

	m.removeUsername(id, username, issuer)
}

func (m *Memory) setUsername(id, username, issuer string) {
	if username == "" {
		return
	}

	key := m.key(issuer, username)

	if slices.Contains(m.lookupUsername[key], id) {
		return
	}

	m.lookupUsername[key] = append(m.lookupUsername[key], id)
}

func (m *Memory) removeUsername(id, username, issuer string) {
	if username == "" {
		return
	}

	key := m.key(issuer, username)

	ids, ok := m.lookupUsername[key]
	if !ok {
		return
	}

	if i := slices.Index(ids, id); i >= 0 {
		ids = slices.Delete(ids, i, i+1)
	}

	if len(ids) == 0 {
		delete(m.lookupUsername, key)
	} else {
		m.lookupUsername[key] = ids
	}
}

func (m *Memory) key(issuer, value string) memoryKey {
	return memoryKey{issuer: issuer, value: value}
}

type memoryKey struct {
	issuer string
	value  string
}

type itemSession struct {
	data      []byte
	id        string
	pid       string
	issuer    string
	username  string
	expires   time.Time
	destroyed bool
}

func (i *itemSession) expired(now time.Time) bool {
	return !i.expires.IsZero() && !i.expires.After(now)
}

func sessionExpires(expiration time.Duration) (expires time.Time) {
	if expiration <= 0 {
		return time.Time{}
	}

	return time.Now().Add(expiration)
}

var (
	_ Provider = (*Memory)(nil)
)
