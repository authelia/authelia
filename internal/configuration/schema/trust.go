package schema

type Trust struct {
	Certificates CertificateTrust `koanf:"certificates"`
}

type CertificateTrust struct {
	Paths                     []string               `koanf:"paths"`
	Certificates              []X509CertificateChain `koanf:"certificates"`
	DisableSystemCertificates bool                   `koanf:"disable_system_certificates"`
	DisableValidationErrors   bool                   `koanf:"disable_validation_errors"`
	DisableValidateNotBefore  bool                   `koanf:"disable_validate_not_before"`
	DisableValidateNotAfter   bool                   `koanf:"disable_validate_not_after"`
}
