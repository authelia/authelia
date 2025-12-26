package authentication

import (
	"sync"

	"github.com/jcmturner/gokrb5/v8/keytab"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type SPNEGOKeytabProvider struct {
	config *schema.SPNEGO
	kt     *keytab.Keytab
	mutex  sync.Mutex
}

func NewSPNEGOKeytabProvider(config *schema.SPNEGO) (*SPNEGOKeytabProvider, error) {
	kt, err := keytab.Load(config.Keytab)
	if err != nil {
		return nil, err
	}

	return &SPNEGOKeytabProvider{
		config: config,
		kt:     kt,
	}, nil
}

func (p *SPNEGOKeytabProvider) GetKeytab() (*keytab.Keytab, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.kt, nil
}

func (p *SPNEGOKeytabProvider) Reload() (reloaded bool, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	keyTabFile := "/etc/krb5.keytab"
	if p.config.Keytab != "" {
		keyTabFile = p.config.Keytab
	}

	kt, err := keytab.Load(keyTabFile)
	if err != nil {
		return false, err
	}

	p.kt = kt

	return true, nil
}
