package authentication

import (
	"sync"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/jcmturner/gokrb5/v8/keytab"
)

type SPNEGOKeytabProvider struct {
	config *schema.SPNEGO
	kt     *keytab.Keytab
	mutex  sync.Mutex
}

func NewSPNEGOKeytabProvider(config *schema.SPNEGO) *SPNEGOKeytabProvider {
	// do not initialize if disabled
	if config.Disable {
		return nil
	}

	keyTabFile := "/etc/krb5.keytab"
	if config.Keytab != "" {
		keyTabFile = config.Keytab
	}

	kt, err := keytab.Load(keyTabFile)
	if err != nil {
		return nil
	}

	return &SPNEGOKeytabProvider{
		config: config,
		kt:     kt,
	}
}

func (p *SPNEGOKeytabProvider) GetKeytab() (*keytab.Keytab, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.kt, nil
}

func (p *SPNEGOKeytabProvider) Reload() (reloaded bool, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.config.Disable {
		return false, nil
	}

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
