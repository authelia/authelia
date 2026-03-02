package schema

type SPNEGO struct {
	Enabled   bool   `koanf:"enabled" json:"enabled" jsonschema:"default=false,title=Enabled" jsonschema_description:"Enables the SPNEGO authentication functionality."`
	Keytab    string `koanf:"keytab" json:"keytab" jsonschema:"title=File Path" jsonschema_description:"The filepath to the kerberos keytab."`
	Principal string `koanf:"principal" json:"principal" jsonschema:"title=Principal" jsonschema_description:"The service principal name."`
	Realm     string `koanf:"realm" json:"realm" jsonschema:"title=Kerberos Realm" jsonschema_description:"The Kerberos realm to use for SPNEGO authentication."`
}
