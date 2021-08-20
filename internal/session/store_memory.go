package session

import (
	"sync"
	"time"

	"github.com/fasthttp/session/v2"

	"github.com/authelia/authelia/v4/internal/utils"
)

// LICENSE: MIT https://github.com/fasthttp/session/blob/master/LICENSE
// SOURCE: https://github.com/fasthttp/session/blob/cd9080042fc350c0b630c401f43e7d5ecee77882/providers/memory/provider.go
// Changes to the original code prior to the first commit were only aesthetic. All other changes are logged via git SCM.

// NewMemoryStore returns a new MemoryStore.
func NewMemoryStore() (store *MemoryStore) {
	p := &MemoryStore{
		db: new(session.Dict),
	}

	return p
}

// MemoryStore is used to store sessions in memory.
type MemoryStore struct {
	db *session.Dict
}

// Get returns the data of the given session id.
func (s *MemoryStore) Get(id []byte) (data []byte, err error) {
	val := s.db.GetBytes(id)
	if val == nil { // Not exist
		return nil, nil
	}

	item := val.(*item)

	return item.data, nil
}

// Save the session data and expiration from the given session id.
func (s *MemoryStore) Save(id, data []byte, expiration time.Duration) (err error) {
	item := acquireItem()
	item.data = data
	item.lastActiveTime = time.Now().UnixNano()
	item.expiration = expiration

	s.db.SetBytes(id, item)

	return nil
}

// Regenerate updates the session id and expiration with the new session id of the given current session id.
func (s *MemoryStore) Regenerate(id, newID []byte, expiration time.Duration) (err error) {
	data := s.db.GetBytes(id)
	if data != nil {
		item := data.(*item)
		item.lastActiveTime = time.Now().UnixNano()
		item.expiration = expiration

		s.db.SetBytes(newID, item)
		s.db.DelBytes(id)
	}

	return nil
}

// Destroy destroys the session from the given id.
func (s *MemoryStore) Destroy(id []byte) (err error) {
	val := s.db.GetBytes(id)
	if val == nil {
		return nil
	}

	s.db.DelBytes(id)
	releaseItem(val.(*item))

	return nil
}

// Count returns the count of stored sessions.
func (s *MemoryStore) Count() (count int) {
	return len(s.db.D)
}

// NeedGC indicates if the GC needs to be run.
func (s *MemoryStore) NeedGC() (needGC bool) {
	return true
}

// GC destroys the expired sessions.
func (s *MemoryStore) GC() (err error) {
	now := time.Now().UnixNano()

	for _, kv := range s.db.D {
		item := kv.Value.(*item)

		if item.expiration == 0 {
			continue
		}

		if now >= (item.lastActiveTime + item.expiration.Nanoseconds()) {
			_ = s.Destroy(utils.StringToByteSlice(kv.Key))
		}
	}

	return nil
}

type item struct {
	data           []byte
	lastActiveTime int64
	expiration     time.Duration
}

var itemPool = &sync.Pool{
	New: func() interface{} {
		return new(item)
	},
}

func acquireItem() *item {
	return itemPool.Get().(*item)
}

func releaseItem(item *item) {
	item.data = item.data[:0]
	item.lastActiveTime = 0
	item.expiration = 0

	itemPool.Put(item)
}
