package cache

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/authelia/authelia/v4/internal/session"
)

// NewMemory returns a new Memory Provider with its lookups initialized.
func NewMemory() (memory *Memory) {
	return &Memory{
		session:        map[string]*itemSession{},
		lookupPublicID: map[string]string{},
		lookupUsername: map[string][]string{},
	}
}

// Memory is a Provider which stores sessions in process memory. Every exported method acquires mu, and the items stored
// in the maps are only ever read or mutated while it is held, which makes the unexported helpers below unsafe to call
// on their own.
type Memory struct {
	mu sync.RWMutex

	session        map[string]*itemSession
	lookupPublicID map[string]string
	lookupUsername map[string][]string
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
	if !ok || item.expired(time.Now()) {
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
	if !ok || item.expired(time.Now()) {
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
		if item, ok := m.session[m.key(issuer, id)]; ok && !item.expired(now) {
			ids = append(ids, id)
		}
	}

	return ids, nil
}

// SessionSave implements the Provider interface.
func (m *Memory) SessionSave(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.key(issuer, id)

	if item, ok := m.session[key]; ok && item.pid != pid {
		delete(m.lookupPublicID, m.key(issuer, item.pid))
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

// SessionSaveData updates the session data. The public id and username are unused as the stored session already records
// them and its lookups are derived from it rather than expiring independently.
// SessionSaveData implements the Provider interface.
func (m *Memory) SessionSaveData(ctx context.Context, issuer, id, _, _ string, expiration time.Duration, data []byte) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.key(issuer, id)

	item, ok := m.session[key]
	if !ok {
		return fmt.Errorf("session not found: %s", key)
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
	if !ok {
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

	// The stored session is the authority on which lookups refer to it. The caller may not know the public id or the
	// username, such as when destroying a session whose data couldn't be decoded, which would otherwise strand them.
	if item, ok := m.session[m.key(issuer, id)]; ok {
		pid, username = item.pid, item.username
	}

	m.delete(id, pid, username, issuer)

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

// delete removes a session and every lookup which refers to it. The caller must hold the write lock.
func (m *Memory) delete(id, pid, username, issuer string) {
	delete(m.session, m.key(issuer, id))
	delete(m.lookupPublicID, m.key(issuer, pid))

	m.removeUsername(id, username, issuer)
}

// setUsername records the id against the username lookup, skipping sessions which have no username and ids which are
// already recorded so the lookup can't grow without bound. The caller must hold the write lock.
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

// removeUsername drops the id from the username lookup, removing the lookup entirely once it's empty. The caller must
// hold the write lock.
func (m *Memory) removeUsername(id, username, issuer string) {
	if username == "" {
		return
	}

	key := m.key(issuer, username)

	ids, ok := m.lookupUsername[key]
	if !ok {
		return
	}

	if ids = removeString(ids, id); len(ids) == 0 {
		delete(m.lookupUsername, key)
	} else {
		m.lookupUsername[key] = ids
	}
}

func (m *Memory) key(values ...string) string {
	return strings.Join(values, ":")
}

type itemSession struct {
	data     []byte
	id       string
	pid      string
	issuer   string
	username string
	expires  time.Time
}

// expired returns true if the session has an expiration which has elapsed. A zero expires value never expires which
// matches the behavior of a zero expiration in the Redis provider.
func (i *itemSession) expired(now time.Time) bool {
	return !i.expires.IsZero() && !i.expires.After(now)
}

// sessionExpires converts a relative expiration into the absolute time the session expires at, matching the way the
// Redis and SQL providers recalculate the expiry on every write. A non-positive expiration never expires.
func sessionExpires(expiration time.Duration) (expires time.Time) {
	if expiration <= 0 {
		return time.Time{}
	}

	return time.Now().Add(expiration)
}

func removeString(s []string, str string) (result []string) {
	for i, v := range s {
		if v == str {
			return append(s[:i], s[i+1:]...)
		}
	}

	return s
}

var (
	_ Provider = (*Memory)(nil)
)
