//nolint:all
package memory

import (
	"sync"
	"time"

	"github.com/savsgio/gotils/strconv"
)

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

// New returns a new memory provider configured
func New(cfg Config) (*Provider, error) {
	p := &Provider{
		config: cfg,
	}

	return p, nil
}

func (p *Provider) getSessionKey(sessionID []byte) string {
	return strconv.B2S(sessionID)
}

// Get returns the data of the given session id
func (p *Provider) Get(id []byte) ([]byte, error) {
	key := p.getSessionKey(id)

	val, found := p.db.Load(key)
	if !found || val == nil { // Not exist
		return nil, nil
	}

	item := val.(*item)

	return item.data, nil
}

// Save saves the session data and expiration from the given session id
func (p *Provider) Save(id, data []byte, expiration time.Duration) error {
	key := p.getSessionKey(id)

	item := acquireItem()
	item.data = data
	item.lastActiveTime = time.Now().UnixNano()
	item.expiration = expiration

	p.db.Store(key, item)

	return nil
}

// Regenerate updates the session id and expiration with the new session id
// of the the given current session id
func (p *Provider) Regenerate(id, newID []byte, expiration time.Duration) error {
	key := p.getSessionKey(id)

	data, found := p.db.LoadAndDelete(key)
	if found && data != nil {
		item := data.(*item)
		item.lastActiveTime = time.Now().UnixNano()
		item.expiration = expiration

		newKey := p.getSessionKey(newID)

		p.db.Store(newKey, item)
	}

	return nil
}

func (p *Provider) destroy(key string) error {
	val, found := p.db.LoadAndDelete(key)
	if !found || val == nil {
		return nil
	}

	releaseItem(val.(*item))

	return nil
}

// Destroy destroys the session from the given id
func (p *Provider) Destroy(id []byte) error {
	key := p.getSessionKey(id)

	return p.destroy(key)
}

// Count returns the total of stored sessions
func (p *Provider) Count() (count int) {
	p.db.Range(func(_, _ interface{}) bool {
		count++

		return true
	})

	return count
}

// NeedGC indicates if the GC needs to be run
func (p *Provider) NeedGC() bool {
	return true
}

// GC destroys the expired sessions
func (p *Provider) GC() error {
	now := time.Now().UnixNano()

	p.db.Range(func(key, value interface{}) bool {
		item := value.(*item)

		if item.expiration == 0 {
			return true
		}

		if now >= (item.lastActiveTime + item.expiration.Nanoseconds()) {
			_ = p.destroy(key.(string))
		}

		return true
	})

	return nil
}
