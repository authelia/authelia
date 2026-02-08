package authentication

type LDAPClient struct {
	LDAPBaseClient

	features LDAPSupportedFeatures
}

func (c *LDAPClient) Features() (features LDAPSupportedFeatures) {
	return c.features
}
