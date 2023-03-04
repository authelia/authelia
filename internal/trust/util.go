package trust

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func loadPEMCertificates(data []byte) (certs []*x509.Certificate, err error) {
	var (
		cert  *x509.Certificate
		block *pem.Block
	)

	for len(data) > 0 {
		if block, data = pem.Decode(data); block == nil {
			if len(certs) == 0 {
				break
			}

			return nil, fmt.Errorf("failed to parse certificate: the file contained no PEM blocks")
		}

		if block.Type != "CERTIFICATE" {
			return nil, fmt.Errorf("failed to parse certificate PEM block: the PEM block is not a certificate, it's a '%s'", block.Type)
		}

		if len(block.Headers) != 0 {
			return nil, fmt.Errorf("failed to parse certificate PEM block: the PEM block has additional unexpected headers")
		}

		if cert, err = x509.ParseCertificate(block.Bytes); err != nil {
			return nil, fmt.Errorf("failed to parse certificate PEM block: %w", err)
		}

		certs = append(certs, cert)
	}

	return certs, nil
}
