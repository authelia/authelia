package cache

import (
	"context"
	"strings"
	"sync"
	"time"

	m "github.com/authelia/authelia/v4/internal/cache/map"
)

func NewMemory() (memory *Memory) {
	return &Memory{
		pool: NewMemoryPools(),
	}
}

type Memory struct {
	session m.Map

	lookupPublicID m.Map
	lookupUsername m.Map

	pool *MemoryPools
}

func (m *Memory) StartupCheck() (err error) {
	return nil
}

func (m *Memory) SessionGet(ctx context.Context, id, issuer string) (data []byte, err error) {
	value, ok := m.session.Load(m.key(issuer, id))
	if !ok || value == nil {
		return nil, nil
	}

	return value.(*itemSession).data, nil
}

func (m *Memory) SessionGetByPublicID(ctx context.Context, pubid, issuer string) (data []byte, err error) {
	value, ok := m.lookupPublicID.Load(m.key(issuer, pubid))
	if !ok || value == nil {
		return nil, nil
	}

	value, ok = m.session.Load(m.key(issuer, value.(*itemPublicID).data))
	if !ok || value == nil {
		return nil, nil
	}

	return value.(*itemSession).data, nil
}

func (m *Memory) SessionGetIDsByUsername(ctx context.Context, username, issuer string) (ids []string, err error) {
	value, ok := m.lookupUsername.Load(m.key(issuer, username))
	if !ok || value == nil {
		return nil, nil
	}

	return value.(*itemUsername).data, nil
}

func (m *Memory) SessionSave(ctx context.Context, id, pubid, username, issuer string, expiration time.Duration, data []byte) (err error) {
	sessionItem := m.pool.acquireSession()
	sessionItem.data = data
	sessionItem.id = id
	sessionItem.pubid = pubid
	sessionItem.issuer = issuer
	sessionItem.username = username
	sessionItem.lastActiveTime = time.Now().UnixNano()
	sessionItem.expiration = expiration

	m.session.Store(m.key(issuer, id), sessionItem)

	publicIDItem := m.pool.acquirePublicID()

	publicIDItem.data = pubid

	m.lookupPublicID.Store(m.key(issuer, pubid), publicIDItem)

	return m.sessionSetUsername(ctx, id, username, issuer)
}

func (m *Memory) SessionSaveData(ctx context.Context, id, issuer string, expiration time.Duration, data []byte) (err error) {
	sessionItem := m.pool.acquireSession()
	sessionItem.data = data
	sessionItem.lastActiveTime = time.Now().UnixNano()
	sessionItem.expiration = expiration

	m.session.Store(m.key(issuer, id), sessionItem)

	return nil
}

func (m *Memory) SessionSetUsername(ctx context.Context, id, username, issuer string) (err error) {
	key := m.key(issuer, id)

	if value, ok := m.session.Load(key); ok && value != nil {
		item := value.(*itemSession)
		item.username = username

		m.session.Store(key, item)
	}

	return m.sessionSetUsername(ctx, id, username, issuer)
}

func (m *Memory) sessionSetUsername(ctx context.Context, id, username, issuer string) (err error) {
	var usernameItem *itemUsername

	value, ok := m.lookupUsername.Load(m.key(issuer, username))
	if value == nil || !ok {
		usernameItem = m.pool.acquireUsername()
	} else {
		usernameItem = value.(*itemUsername)
	}

	usernameItem.data = append(usernameItem.data, id)

	m.lookupUsername.Store(m.key(issuer, username), usernameItem)

	return nil
}

func (m *Memory) SessionChangeID(ctx context.Context, oldID, id, pubid, username, issuer string, expiration time.Duration) (err error) {
	key := m.key(issuer, oldID)

	if value, loaded := m.session.LoadAndDelete(key); loaded && value != nil {
		item := value.(*itemSession)
		item.lastActiveTime = time.Now().UnixNano()
		item.expiration = expiration

		m.session.Store(m.key(issuer, id), item)
	}

	key = m.key(issuer, pubid)

	if value, loaded := m.lookupPublicID.LoadAndDelete(key); loaded && value != nil {
		item := value.(*itemPublicID)
		item.data = id

		m.session.Store(key, item)
	}

	key = m.key(issuer, username)

	if value, loaded := m.lookupPublicID.LoadAndDelete(key); loaded && value != nil {
		item := value.(*itemUsername)

		item.data = replaceOrAppendString(item.data, oldID, id)

		m.session.Store(key, item)
	}

	return nil
}

func (m *Memory) SessionDelete(ctx context.Context, id, pubid, username, issuer string) (err error) {
	m.destroySession(m.key(issuer, id))
	m.destroyLookupPublicID(m.key(issuer, pubid))
	m.destroyLookupUsername(m.key(issuer, username), id)

	return nil
}

func (m *Memory) SessionGarbageCollectionRequired(ctx context.Context) bool {
	return true
}

func (m *Memory) SessionGarbageCollection(ctx context.Context) error {
	now := time.Now().UnixNano()

	m.session.Range(func(k, v any) bool {
		item := v.(*itemSession)

		if item.expiration == 0 {
			return true
		}

		if now >= (item.lastActiveTime + item.expiration.Nanoseconds()) {
			key := k.(string)

			m.destroySession(key)
			m.destroyLookupPublicID(m.key(item.issuer, item.pubid))
			m.destroyLookupUsername(m.key(item.issuer, item.username), item.id)
		}

		return true
	})

	return nil
}

func (m *Memory) destroySession(key string) {
	if value, loaded := m.session.LoadAndDelete(key); loaded && value != nil {
		m.pool.releaseSession(value.(*itemSession))
	}
}

func (m *Memory) destroyLookupPublicID(key string) {
	if value, loaded := m.lookupPublicID.LoadAndDelete(key); loaded && value != nil {
		m.pool.releasePublicID(value.(*itemPublicID))
	}
}

func (m *Memory) destroyLookupUsername(key, id string) {
	if value, ok := m.lookupUsername.Load(key); ok && value != nil {
		usernameItem := value.(*itemUsername)

		usernameItem.data = removeString(usernameItem.data, id)

		if len(usernameItem.data) == 0 {
			m.lookupUsername.Delete(key)
		} else {
			m.lookupUsername.Store(key, usernameItem)
		}
	}
}

func (m *Memory) key(values ...string) string {
	return strings.Join(values, ":")
}

func NewMemoryPools() *MemoryPools {
	return &MemoryPools{
		session: &sync.Pool{
			New: func() any {
				return new(itemSession)
			},
		},
		lookupPublicID: &sync.Pool{
			New: func() any {
				return new(itemPublicID)
			},
		},
		lookupUsername: &sync.Pool{
			New: func() any {
				return new(itemUsername)
			},
		},
	}
}

type MemoryPools struct {
	session        *sync.Pool
	lookupPublicID *sync.Pool
	lookupUsername *sync.Pool
}

func (m *MemoryPools) acquireSession() *itemSession {
	return m.session.Get().(*itemSession)
}

func (m *MemoryPools) acquirePublicID() *itemPublicID {
	return m.lookupPublicID.Get().(*itemPublicID)
}

func (m *MemoryPools) acquireUsername() *itemUsername {
	return m.lookupUsername.Get().(*itemUsername)
}

func (m *MemoryPools) releaseSession(item *itemSession) {
	item.data = item.data[:0]
	item.id = ""
	item.pubid = ""
	item.issuer = ""
	item.username = ""
	item.lastActiveTime = 0
	item.expiration = 0

	m.session.Put(item)
}

func (m *MemoryPools) releasePublicID(item *itemPublicID) {
	item.data = item.data[:0]

	m.lookupPublicID.Put(item)
}

func (m *MemoryPools) releaseUsername(item *itemUsername) {
	item.data = item.data[:0]

	m.lookupUsername.Put(item)
}

type itemSession struct {
	data           []byte
	id             string
	pubid          string
	issuer         string
	username       string
	lastActiveTime int64
	expiration     time.Duration
}

type itemPublicID struct {
	data string
}

type itemUsername struct {
	data []string
}

func removeString(s []string, str string) (result []string) {
	for i, v := range s {
		if v == str {
			return append(s[:i], s[i+1:]...)
		}
	}

	return s
}

func replaceOrAppendString(s []string, match, value string) (result []string) {
	for i, v := range s {
		if v == match {
			s[i] = value

			return s
		}
	}

	return append(s, value)
}

var (
	_ Provider = (*Memory)(nil)
)
