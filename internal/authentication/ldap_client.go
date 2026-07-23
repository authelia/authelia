package authentication

// LDAPClient is a wrapped LDAPBaseClient which also has discovery information.
type LDAPClient struct {
	LDAPBaseClient

	discovery LDAPDiscovery
}

// Discovery implements LDAPExtendedClient and returns the discovery information for this client.
func (c *LDAPClient) Discovery() (discovery LDAPDiscovery) {
	return c.discovery
}

func (c *LDAPClient) setDiscovery(discovery LDAPDiscovery) {
	c.discovery = discovery
}

type ldapDiscoverySetter interface {
	setDiscovery(discovery LDAPDiscovery)
}
