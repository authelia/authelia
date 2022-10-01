package schema

import (
	"crypto"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAddressFromString(t *testing.T) {
	testCases := []struct {
		name                                          string
		have                                          string
		expected                                      *Address
		expectedHostPort, expectedString, expectedErr string
	}{
		{"ShouldParseBasicAddress", "tcp://0.0.0.0:9091", &Address{true, "tcp", net.ParseIP("0.0.0.0"), 9091}, "0.0.0.0:9091", "tcp://0.0.0.0:9091", ""},
		{"ShouldParseEmptyAddress", "", &Address{true, "tcp", net.ParseIP("0.0.0.0"), 0}, "0.0.0.0:0", "tcp://0.0.0.0:0", ""},
		{"ShouldParseAddressMissingScheme", "0.0.0.0:9091", &Address{true, "tcp", net.ParseIP("0.0.0.0"), 9091}, "0.0.0.0:9091", "tcp://0.0.0.0:9091", ""},
		{"ShouldParseAddressMissingPort", "tcp://0.0.0.0", &Address{true, "tcp", net.ParseIP("0.0.0.0"), 0}, "0.0.0.0:0", "tcp://0.0.0.0:0", ""},
		{"ShouldNotParseUnknownScheme", "a://0.0.0.0", nil, "", "", "could not parse scheme for address 'a://0.0.0.0': scheme 'a' is not valid, expected to be one of 'tcp://', 'udp://'"},
		{"ShouldNotParseInvalidPort", "tcp://0.0.0.0:abc", nil, "", "", "could not parse string 'tcp://0.0.0.0:abc' as address: expected format is [<scheme>://]<ip>[:<port>]: parse \"tcp://0.0.0.0:abc\": invalid port \":abc\" after host"},
		{"ShouldNotParseInvalidIP", "tcp://example.com:9091", nil, "", "", "could not parse ip for address 'tcp://example.com:9091': example.com does not appear to be an IP"},
		{"ShouldNotParseInvalidAddress", "@$@#%@#$@@", nil, "", "", "could not parse string '@$@#%@#$@@' as address: expected format is [<scheme>://]<ip>[:<port>]: parse \"tcp://@$@#%@#$@@\": invalid URL escape \"%@#\""},
		{"ShouldNotParseInvalidAddressWithScheme", "tcp://@$@#%@#$@@", nil, "", "", "could not parse string 'tcp://@$@#%@#$@@' as address: expected format is [<scheme>://]<ip>[:<port>]: parse \"tcp://@$@#%@#$@@\": invalid URL escape \"%@#\""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, actualErr := NewAddressFromString(tc.have)

			if len(tc.expectedErr) != 0 {
				assert.EqualError(t, actualErr, tc.expectedErr)
			} else {
				assert.Nil(t, actualErr)

				assert.Equal(t, actual.HostPort(), tc.expectedHostPort)
				assert.Equal(t, actual.String(), tc.expectedString)
			}

			assert.Equal(t, tc.expected, actual)

			if actual != nil {
				assert.True(t, actual.Valid())
				assert.NotEmpty(t, actual.String())
				assert.NotEmpty(t, actual.HostPort())
			}
		})
	}
}

func TestNewX509CertificateChain(t *testing.T) {
	testCases := []struct {
		name             string
		have             string
		thumbprintSHA256 string
		err              string
	}{
		{"ShouldParseCertificate", x509CertificateRSA,
			"f3cf205d419b46c01315d52cadd5f28095cea3af894aca2bd80e630f33f7fa0d", ""},
		{"ShouldParseCertificateChain", x509CertificateRSA + "\n" + x509CACertificateRSA,
			"f3cf205d419b46c01315d52cadd5f28095cea3af894aca2bd80e630f33f7fa0d", ""},
		{"ShouldNotParseInvalidCertificate", x509CertificateRSAInvalid, "",
			"the PEM data chain contains an invalid certificate: x509: malformed certificate"},
		{"ShouldNotParseInvalidCertificateBlock", x509CertificateRSAInvalidBlock, "", "invalid PEM block"},
		{"ShouldNotParsePrivateKey", x509PrivateKeyRSA, "",
			"the PEM data chain contains a RSA PRIVATE KEY but only certificates are expected"},
		{"ShouldNotParseEmptyPEMBlock", x509CertificateEmpty, "", "invalid PEM block"},
		{"ShouldNotParseEmptyData", "", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewX509CertificateChain(tc.have)

			switch len(tc.err) {
			case 0:
				switch len(tc.have) {
				case 0:
					assert.Nil(t, actual)
				default:
					assert.NotNil(t, actual)

					assert.Equal(t, tc.thumbprintSHA256, fmt.Sprintf("%x", actual.Thumbprint(crypto.SHA256)))
					assert.True(t, actual.HasCertificates())
				}
				assert.NoError(t, err)
			default:
				assert.Nil(t, actual)
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

var (
	x509CertificateRSA = `-----BEGIN CERTIFICATE-----
MIIC5jCCAc6gAwIBAgIRAK4Sj7FiN6PXo/urPfO4E7owDQYJKoZIhvcNAQELBQAw
EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNNzAwMTAxMDAwMDAwWhcNNzEwMTAxMDAw
MDAwWjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAPKv3pSyP4ozGEiVLJ14dIWFCEGEgq7WUMI0SZZqQA2ID0L59U/Q
/Usyy7uC9gfMUzODTpANtkOjFQcQAsxlR1FOjVBrX5QgjSvXwbQn3DtwMA7XWSl6
LuYx2rBYSlMSN5UZQm/RxMtXfLK2b51WgEEYDFi+nECSqKzR4R54eOPkBEWRfvuY
91AMjlhpivg8e4JWkq4LVQUKbmiFYwIdK8XQiN4blY9WwXwJFYs5sQ/UYMwBFi0H
kWOh7GEjfxgoUOPauIueZSMSlQp7zqAH39N0ZSYb6cS0Npj57QoWZSY3ak87ebcR
Nf4rCvZLby7LoN7qYCKxmCaDD3x2+NYpWH8CAwEAAaM1MDMwDgYDVR0PAQH/BAQD
AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
AQELBQADggEBAHSITqIQSNzonFl3DzxHPEzr2hp6peo45buAAtu8FZHoA+U7Icfh
/ZXjPg7Xz+hgFwM/DTNGXkMWacQA/PaNWvZspgRJf2AXvNbMSs2UQODr7Tbv+Fb4
lyblmMUNYFMCFVAMU0eIxXAFq2qcwv8UMcQFT0Z/35s6PVOakYnAGGQjTfp5Ljuq
wsdc/xWmM0cHWube6sdRRUD7SY20KU/kWzl8iFO0VbSSrDf1AlEhnLEkp1SPaxXg
OdBnl98MeoramNiJ7NT6Jnyb3zZ578fjaWfThiBpagItI8GZmG4s4Ovh2JbheN8i
ZsjNr9jqHTjhyLVbDRlmJzcqoj4JhbKs6/I=
-----END CERTIFICATE-----`

	x509CACertificateRSA = `-----BEGIN CERTIFICATE-----
MIIDBDCCAeygAwIBAgIRALJsPg21kA0zY4F1wUCIuoMwDQYJKoZIhvcNAQELBQAw
EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNNzAwMTAxMDAwMDAwWhcNNzEwMTAxMDAw
MDAwWjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAMXHBvVxUzYk0u34/DINMSF+uiOekKOAjOrC6Mi9Ww8ytPVO7t2S
zfTvM+XnEJqkFQFgimERfG/eGhjF9XIEY6LtnXe8ATvOK4nTwdufzBaoeQu3Gd50
5VXr6OHRo//ErrGvFXwP3g8xLePABsi/fkH3oDN+ztewOBMDzpd+KgTrk8ysv2ou
kNRMKFZZqASvCgv0LD5KWvUCnL6wgf1oTXG7aztduA4oSkUP321GpOmBC5+5ElU7
ysoRzvD12o9QJ/IfEaulIX06w9yVMo60C/h6A3U6GdkT1SiyTIqR7v7KU/IWd/Qi
Lfftcj91VhCmJ73Meff2e2S2PrpjdXbG5FMCAwEAAaNTMFEwDgYDVR0PAQH/BAQD
AgKkMA8GA1UdJQQIMAYGBFUdJQAwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU
Z7AtA3mzFc0InSBA5fiMfeLXA3owDQYJKoZIhvcNAQELBQADggEBAEE5hm1mtlk/
kviCoHH4evbpw7rxPxDftIQlqYTtvMM4eWY/6icFoSZ4fUHEWYyps8SsPu/8f2tf
71LGgZn0FdHi1QU2H8m0HHK7TFw+5Q6RLrLdSyk0PItJ71s9en7r8pX820nAFEHZ
HkOSfJZ7B5hFgUDkMtVM6bardXAhoqcMk4YCU96e9d4PB4eI+xGc+mNuYvov3RbB
D0s8ICyojeyPVLerz4wHjZu68Z5frAzhZ68YbzNs8j2fIBKKHkHyLG1iQyF+LJVj
2PjCP+auJsj6fQQpMGoyGtpLcSDh+ptcTngUD8JsWipzTCjmaNqdPHAOYmcgtf4b
qocikt3WAdU=
-----END CERTIFICATE-----`

	x509CertificateRSAInvalid = `-----BEGIN CERTIFICATE-----
mIIC5jCCAc6gAwIBAgIRAK4Sj7FiN6PXo/urpfO4E7owDQYJKoZIhvcNAQELBQAw
EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNNzAwMTAxMDAwMDAwWhcNNzEwMTAxMDAw
MDAwWjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAPKv3pSyP4ozGEiVLJ14dIwFCEGEgq7WUMI0SZZqQA2ID0L59U/Q
/Usyy7uC9gfMUzODTpANtkOjFQcQAsxlR1FOjVBrX5QgjSvXwbQn3DtwMA7XWSl6
LuYx2rBYSlMSN5UZQm/RxMtXfLK2b51WgEEYDFi+nECSqKzR4R54eOPkBEWRfvuY
91AMjlhpivg8e4JWkq4LVQUKbmiFYwIdK8XQiN4blY9WwXwJFYs5sQ/UYMwBFi0H
kWOh7GEjfxgoUOPauIueZSMSlQp7zqAH39n0ZSYb6cS0Npj57QoWZSY3ak87ebcR
Nf4rCvZLby7LoN7qYCKxmCaDD3x2+NYpWH8CAwEAAaM1MDMwDgYDVR0PAQH/BAQD
AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
AQELBQADggEBAHSITqIQSNzonFl3DzxHPEzr2hp6peo45buAAtu8FZHoA+U7Icfh
/ZXjPg7Xz+hgFwM/DTNGXkMWacQA/PaNWvZspgRJf2AXvNbMSs2UQODr7Tbv+Fb4
lyblmMUNYFMCFVAMU0eIxXAFq2qcwv8UMcQFT0Z/35s6PVOakYnAGGQjTfp5Ljuq
wsdc/xWmM0cHWube6sdRRUD7SY20KU/kWzl8iFO0VbSSrDf1AlEhnLEkp1SPaxXg
OdBnl98MeoramNiJ7NT6Jnyb3zZ578fjaWfThiBpagItI8GZmG4s4Ovh2JbheN8i
ZsjNr9jqHTjhyLVbDRlmJzcqoj4JhbKs6/I=
-----END CERTIFICATE-----`

	x509CertificateRSAInvalidBlock = `-----BEGIN CERTIFICATE-----
MIIC5jCCAc6gAwIBAgIRAK4Sj7FiN6PXo/urPfO4E7owDQYJKoZIhvcNAQELBQAw
EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNNzAwMTAxMDAwMDAwWhcNNzEwMTAxMDAw
MDAwWjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAPKv3pSyP4ozGEiVLJ14dIWFCEGEgq7WUMI0SZZqQA2ID0L59U/Q
/Usyy7uC9gfMUzODTpANtkOjFQcQAsxlR1FOjVBrX5QgjSvXwbQn3DtwMA7XWSl6
LuYx2rBYSlMSN5UZQm/RxMtXf^K2b51WgEEYDFi+nECSqKzR4R54eOPkBEWRfvuY
91AMjlhpivg8e4JWkq4LVQUKbmiFYwIdK8XQiN4blY9WwXwJFYs5sQ/UYMwBFi0H
kWOh7GEjfxgoUOPauIueZSMSlQp7zqAH39N0ZSYb6cS0Npj57QoWZSY3ak87ebcR
Nf4rCvZLby7LoN7qYCKxmCaDD3x2+NYpWH8CAwEAAaM1MDMwDgYDVR0PAQH/BAQD
AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
AQELBQADggEBAHSITqIQSNzonFl3DzxHPEzr2hp6peo45buAAtu8FZHoA+U7Icfh
/ZXjPg7Xz+hgFwM/DTNGXkMWacQA/PaNWvZspgRJf2AXvNbMSs2UQODr7Tbv+Fb4
lyblmMUNYFMCFVAMU0eIxXAFq2qcwv8UMcQFT0Z/35s6PVOakYnAGGQjTfp5Ljuq
wsdc/xWmM0cHWube6sdRRUD7SY20KU/kWzl8iFO0VbSSrDf1AlEhnLEkp1SPaxXg
OdBnl98MeoramNiJ7NT6Jnyb3zZ578fjaWfThiBpagItI8GZmG4s4Ovh2JbheN8i
ZsjNr9jqHTjhyLVbDRlmJzcqoj4JhbKs6/I=
-----END CERTIFICATE-----`

	x509CertificateEmpty = `-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----`

	x509PrivateKeyRSA = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA8q/elLI/ijMYSJUsnXh0hYUIQYSCrtZQwjRJlmpADYgPQvn1
T9D9SzLLu4L2B8xTM4NOkA22Q6MVBxACzGVHUU6NUGtflCCNK9fBtCfcO3AwDtdZ
KXou5jHasFhKUxI3lRlCb9HEy1d8srZvnVaAQRgMWL6cQJKorNHhHnh44+QERZF+
+5j3UAyOWGmK+Dx7glaSrgtVBQpuaIVjAh0rxdCI3huVj1bBfAkVizmxD9RgzAEW
LQeRY6HsYSN/GChQ49q4i55lIxKVCnvOoAff03RlJhvpxLQ2mPntChZlJjdqTzt5
txE1/isK9ktvLsug3upgIrGYJoMPfHb41ilYfwIDAQABAoIBAQDTOdFf2JjHH1um
aPgRAvNf9v7Nj5jytaRKs5nM6iNf46ls4QPreXnMhqSeSwj6lpNgBYxOgzC9Q+cc
Y4ob/paJJPaIJTxmP8K/gyWcOQlNToL1l+eJ20eQoZm23NGr5fIsunSBwLEpTrdB
ENqqtcwhW937K8Pxy/Q1nuLyU2bc6Tn/ivLozc8n27dpQWWKh8537VY7ancIaACr
LJJLYxKqhQpjtBWAyCDvZQirnAOm9KnvIHaGXIswCZ4Xbsu0Y9NL+woARPyRVQvG
jfxy4EmO9s1s6y7OObSukwKDSNihAKHx/VIbvVWx8g2Lv5fGOa+J2Y7o9Qurs8t5
BQwMTt0BAoGBAPUw5Z32EszNepAeV3E2mPFUc5CLiqAxagZJuNDO2pKtyN29ETTR
Ma4O1cWtGb6RqcNNN/Iukfkdk27Q5nC9VJSUUPYelOLc1WYOoUf6oKRzE72dkMQV
R4bf6TkjD+OVR17fAfkswkGahZ5XA7j48KIQ+YC4jbnYKSxZTYyKPjH/AoGBAP1i
tqXt36OVlP+y84wWqZSjMelBIVa9phDVGJmmhz3i1cMni8eLpJzWecA3pfnG6Tm9
ze5M4whASleEt+M00gEvNaU9ND+z0wBfi+/DwJYIbv8PQdGrBiZFrPhTPjGQUldR
lXccV2meeLZv7TagVxSi3DO6dSJfSEHyemd5j9mBAoGAX8Hv+0gOQZQCSOTAq8Nx
6dZcp9gHlNaXnMsP9eTDckOSzh636JPGvj6m+GPJSSbkURUIQ3oyokMNwFqvlNos
fTaLhAOfjBZI9WnDTTQxpugWjphJ4HqbC67JC/qIiw5S6FdaEvGLEEoD4zoChywZ
9oGAn+fz2d/0/JAH/FpFPgsCgYEAp/ipZgPzziiZ9ov1wbdAQcWRj7RaWnssPFpX
jXwEiXT3CgEMO4MJ4+KWIWOChrti3qFBg6i6lDyyS6Qyls7sLFbUdC7HlTcrOEMe
rBoTcCI1GqZNlqWOVQ65ZIEiaI7o1vPBZo2GMQEZuq8mDKFsOMThvvTrM5cAep84
n6HJR4ECgYABWcbsSnr0MKvVth/inxjbKapbZnp2HUCuw87Ie5zK2Of/tbC20wwk
yKw3vrGoE3O1t1g2m2tn8UGGASeZ842jZWjIODdSi5+icysQGuULKt86h/woz2SQ
27GoE2i5mh6Yez6VAYbUuns3FcwIsMyWLq043Tu2DNkx9ijOOAuQzw==
-----END RSA PRIVATE KEY-----`
)
