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

	keyTabFile := "/etc/krb5.keytab"
	if config.Keytab != "" {
		keyTabFile = config.Keytab
	}

	kt, err := keytab.Load(keyTabFile)
	if err != nil {
		panic("unable to load SPNEGO keytab file: " + err.Error())
	}

	return &SPNEGOKeytabProvider{
		config: config,
		kt:     kt,
	}
}

func (p *SPNEGOKeytabProvider) GetKeytab() (*keytab.Keytab, error) {
	return p.kt, nil
}

func (p *SPNEGOKeytabProvider) Reload() (reloaded bool, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	kt, err := keytab.Load(p.config.Keytab)
	if err != nil {
		return false, err
	}

	p.kt = kt

	return true, nil
}
