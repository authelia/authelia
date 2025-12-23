package schema

type SPNEGO struct {
	Disable   bool   `koanf:"disable" json:"disable" jsonschema:"default=false,title=Disable" jsonschema_description:"Disables the SPNEGO authentication functionality."`
	Keytab    string `koanf:"keytab" json:"keytab" jsonschema:"title=File Path" jsonschema_description:"The filepath to the kerberos keytab."`
	Principal string `koanf:"principal" json:"principal" jsonschema:"title=Principal" jsonschema_description:"The service principal name."`
	Realm     string `koanf:"realm" json:"realm" jsonschema:"title=Kerberos Realm" jsonschema_description:"The Kerberos realm to use for SPNEGO authentication."`
}
