package trust

import (
	"crypto/x509"
	"encoding/pem"
)

func loadPEMCertificates(data []byte) (certs []*x509.Certificate) {
	var (
		cert *x509.Certificate
		err  error
	)

	for len(data) > 0 {
		var block *pem.Block

		block, data = pem.Decode(data)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}

		if cert, err = x509.ParseCertificate(block.Bytes); err != nil {
			continue
		}

		certs = append(certs, cert)
	}

	return certs
}
