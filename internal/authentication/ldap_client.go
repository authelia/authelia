package authentication

type LDAPClient struct {
	LDAPBaseClient

	features LDAPSupportedFeatures
}

func (c *LDAPClient) SetFeatures(features LDAPSupportedFeatures) {
	c.features = features
}

func (c *LDAPClient) Features() LDAPSupportedFeatures {
	return c.features
}
