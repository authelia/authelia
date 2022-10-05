package oidc

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/ory/fosite/token/jwt"
	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

var rsaKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAsGhdSoSPEFYnpiUkILxdixNsSc3LLMsfs/oZeDo6e1XX0SoL
buNDqxCbhVdU+uduF6S3qALFJ6Fb+YCHR1PxTP/TJMN7QacN1CT6PPxxPQBME1c5
0o/4Ce6I7jxp4I/yQfSt8rEJsOUgd7P3IZkOtEkwj8Pc5Yo97fMNwv0RgENwAx7P
6H+zc+Gv4+RA2ZgwA4fRetpHbZ+BCqM9PAeP4ejjck7xW4Ts6zh2x5a+8EV3Y7h1
FUJUa/Ku8Z+jtMPGl79V1c8VOMK+RfYqH8VTUYqPR/RAt/3RzwSboLLIBZ2G7vdw
rbrb6D4EdofflsFr9oEFK015TXzw8wZ8f+TC1QIDAQABAoIBAHh8r8tnwrcrwSTv
lT9uqI2HFJ1MHAtaKSsqFR9S1gLLXP6Vsv1n6B3819w5C+fbDgzEClSMn9Azm7hM
GxmSsesfiD1B7vumeAF/yBLDcSxy+YL0PqAciRNvgsMFLGOerZ5y2iQ62x2hQY3A
s3iVK7/jhXGMG2IEC8xsB+g4JS3uvnN7Zok0L7t0Bf9b/hhjgDLfRsveVTnLrWTK
yiGYnqpI1gOb/4lBRBfJYmmdBKucxobVPY3pQBVluL00uK/T2YZUZGApRmOaOjDo
O2wsmaBileFmOyT5QjKSJhHApnh7lZHUi90TsnwpaSSekid2G+ym9Ducm2xRugiH
kAzfcIECgYEAxJIZs2leUILPhamXhhh6+LF1ZdLvJXpMGvoye2JUAtcA6BGnGVXf
c6M2srl3NvoFWOZ/QVcW4l78QltpP9HHRacwzsPiHaJeZ6jypqcgrpdqY0RNbnBy
pltg6PUdzJF2v6+qxl495ktpvHL1b2koRRQ4l2iaybOj9AUEKaSkgkUCgYEA5b22
5XPBszHV+tqbcHvEJJUjNrxl4cUEuxrnykFctcfS1LPEHNm1FznzxpPLSV46QB6y
6iGHhg1uw/Vb7mO7g93pr8aSgpd/eJn1EKynIHAStSLpbGFkTqX0VYdo92vEVCuO
bPgdGQa8npcPCtW3a8uPvceLAxhWPxwwDPF3j1ECgYA5juZDqJjbTlJhuxUJSOXJ
KH1NwYQRH0xlodliU5px8m7rhS++tWxmueXsU25bBL7IF0Yv5cZnppSLAaVB8LU+
6gPap3TwZHjsNYZH0iw5s1CNnJRnwDlyCNPJouyE2BtDabbBuxq48mAVtDu10b7e
61reytx4L0fUzhs37mPVWQKBgE+tSSYwzEfii8yxTmFnezIYyxqrokX3t1lQzny1
yHp+79627df3pTeF8Ma48TLjzB36x6AturvCIt0xVg3KZvkn2GkO3DcQZbQk6Po8
dsXoOIS7s+rTqB8irSeQi9XreS6b4IxoTGcmj/oMd7oRPsjS02pFLzAGm7rNgtiq
UgXRAoGAPQdGcMLvMq/7l8NzWFzgKRxICfUEGuywJPRvl9sT7sDQ58RqowV2cCPz
+fnB3xCWzzsVpWQm59PWC52hwwNAk6096ZWQZi+3RDYFpHipYdgxgSPMGFM3NSv9
u9L21RKMjbmxWkJsxk9gwMdJISnCW1vQzzhntARKP9B7b8RVzuY=
-----END RSA PRIVATE KEY-----`

var rsaChain = `-----BEGIN CERTIFICATE-----
MIIDBzCCAe+gAwIBAgIQOTV4eUT2SNi7wvqkwtOq4DANBgkqhkiG9w0BAQsFADAT
MREwDwYDVQQKEwhBdXRoZWxpYTAeFw0yMjA5MDkxMTIzMDZaFw0yMzA5MDkxMTIz
MDZaMBMxETAPBgNVBAoTCEF1dGhlbGlhMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAsGhdSoSPEFYnpiUkILxdixNsSc3LLMsfs/oZeDo6e1XX0SoLbuND
qxCbhVdU+uduF6S3qALFJ6Fb+YCHR1PxTP/TJMN7QacN1CT6PPxxPQBME1c50o/4
Ce6I7jxp4I/yQfSt8rEJsOUgd7P3IZkOtEkwj8Pc5Yo97fMNwv0RgENwAx7P6H+z
c+Gv4+RA2ZgwA4fRetpHbZ+BCqM9PAeP4ejjck7xW4Ts6zh2x5a+8EV3Y7h1FUJU
a/Ku8Z+jtMPGl79V1c8VOMK+RfYqH8VTUYqPR/RAt/3RzwSboLLIBZ2G7vdwrbrb
6D4EdofflsFr9oEFK015TXzw8wZ8f+TC1QIDAQABo1cwVTAOBgNVHQ8BAf8EBAMC
BaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDAYDVR0TAQH/BAIwADAgBgNVHREEGTAX
ghVhdXRoLmphbWVzZWxsaW90dC5kZXYwDQYJKoZIhvcNAQELBQADggEBAHthEfao
VG1VXc0cOwiZ+1NzCTRGwTHpQKyB1VM03919GsJIV64Mu0k5GZH48D3Pdgr0wQPx
QSojvxn3vhVFDO0rbDFVAAIO7NIvQ/tjCJiaokXfsPXnwBFWvR8TPLbcnk1X4GNQ
e6JhSKmytWbBNYnnEKxakTv7iKWUp23QGidKZ91nYjPXG18HDZewCCWLjb+Ii4Hv
BzDV1o9V3EclYZFBTyxXMFFcFXp/9GHSicuiVlb0p3iON7woU9ZNvZclDMbzbSn5
D647uyxQoA0/SPNPRT5KF/6VX4eOEXHRT/hTPWJ+BEpLk4lvV/OL1Rvrdh/Je2PT
9bdlrRWemg09w3I=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDAzCCAeugAwIBAgIQMgnu+11Ghv07dLggoKdDjTANBgkqhkiG9w0BAQsFADAT
MREwDwYDVQQKEwhBdXRoZWxpYTAeFw0yMjA5MDkxMTIxMzlaFw0yMzA5MDkxMTIx
MzlaMBMxETAPBgNVBAoTCEF1dGhlbGlhMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAt4sMpm3knPGdSmv21mBIuIbceJq9bBDVwsozjn6b1u79CZakcoXD
ntrMxC8hS073k++qRG76ayuRYrpDOrxtoXhntNmyfSWCHFvHkBBoRNlgb3ZfpZaz
F28xLNrPNVR5eMBEQLQebvbCGQ0pBujC3YK7aLJy6H4itbm2AaaRPoU/kQelWn/3
UvF08PxSMtynET3DUyDAHd+PEwXZ4/5bB0FtAJ6dtabf4z9pfGO1qfKlqTjZ/xUq
JuPI0KXBUmdo1i48L7tYSVMe5BkPx80ee/6AwQxIXkCdQXUt86xlnFlOHdDTuw0S
VrQ4gOrT1HAvZSO0bqlEfSg9/lbAiaJcfQIDAQABo1MwUTAOBgNVHQ8BAf8EBAMC
AqQwDwYDVR0lBAgwBgYEVR0lADAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBSS
KqG/9vigmQuJ+KKdF23nkiTMITANBgkqhkiG9w0BAQsFAAOCAQEAZmnx3fiiv0Of
jtyNxaiJGYxry5e6w1E39fS97SktdIYGk6Qc3adAjTTodFl+IpH3mrSaL3NrPpjT
2lg4c5vgYbXSthHgOnWP4KDvwON8x+ddsXFIJpI5AgnAsBpW2sf59ed40LFiHNlV
OO/A8VGh3HYMRruJH3EfOlTlIzM3TO8FhKhLSedVZoxESoPK00YSH8pu1688R2yy
VcLavxIo7xkvt6O86XTakmz+l0CAw05/F6TKBLtORJ9I8OKzttvWNYTdYjZvsjZr
wmbLPb4hx1gQzMMKw3e9hrJaQdLt4yny4R2+gdKxPFaCtUITWGn9wMxNF5ZtMFRS
UBCxMdLwYA==
-----END CERTIFICATE-----`

var ecdsaKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIGobpXPlTOLHWzXHQkTj/0JUiJBtBXXtLTduelrkMDocoAoGCCqGSM49
AwEHoUQDQgAEIUOsTVUb22uzPqpL4E7ryWN0uBpVx+hJ21OenVOmxhemkwLG914P
RWLPAG+cvdQuQbkcYQku9XP3j9Krq45I7w==
-----END EC PRIVATE KEY-----`

var ecdsaChain = `-----BEGIN CERTIFICATE-----
MIIBVzCB/6ADAgECAhBYpVqewS/0B18536a2xKHOMAoGCCqGSM49BAMCMBMxETAP
BgNVBAoTCEF1dGhlbGlhMB4XDTIyMTAwNDExNDM0OVoXDTIzMTAwNDExNDM0OVow
EzERMA8GA1UEChMIQXV0aGVsaWEwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQh
Q6xNVRvba7M+qkvgTuvJY3S4GlXH6EnbU56dU6bGF6aTAsb3Xg9FYs8Ab5y91C5B
uRxhCS71c/eP0qurjkjvozUwMzAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYI
KwYBBQUHAwEwDAYDVR0TAQH/BAIwADAKBggqhkjOPQQDAgNHADBEAiBW8d25RLi+
POjdYIPrYF1Ja7O62mdKI6rprTSoOEl8wgIgNmdn9UzeTu+hckRgrlCSXOag3Qyf
ukVmuZ8tBvbXmMg=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIBeDCCAR6gAwIBAgIRAIPvuXqjyuO7vMLd/aLj5b4wCgYIKoZIzj0EAwIwEzER
MA8GA1UEChMIQXV0aGVsaWEwHhcNMjIxMDA0MTE0MzM1WhcNMjMxMDA0MTE0MzM1
WjATMREwDwYDVQQKEwhBdXRoZWxpYTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IA
BGKqfuoRYrxmpinisf4eIBrKKUbo6ZsMU0igwmeCa3NhNTly9IKHshMk2ryzNTlw
iqSm5e5X3JvZwFffBkoE3oajUzBRMA4GA1UdDwEB/wQEAwICpDAPBgNVHSUECDAG
BgRVHSUAMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFPro5u7TsVsAxYtdefJ3
fC+M/6DjMAoGCCqGSM49BAMCA0gAMEUCIQD8Of7kdiRwzSEDAgbz4TLfOEll+CA/
KswoyJc0gAkKvwIgK1oJnkf0fgMuG+a7J/Y/GygutVQlm1RUlmSUYmGRN1w=
-----END CERTIFICATE-----`

func MustParseX509CertificateChain(datas ...string) *schema.X509CertificateChain {
	chain, err := schema.NewX509CertificateChain(BuildChain(datas...))
	if err != nil {
		panic(err)
	}

	return chain
}

func BuildChain(pems ...string) string {
	buf := bytes.Buffer{}

	for i, data := range pems {
		if i != 0 {
			buf.WriteString("\n")
		}

		buf.WriteString(data)
	}

	return buf.String()
}

func MustParseRSAPrivateKey(data string) *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(data))
	if block == nil || block.Bytes == nil || len(block.Bytes) == 0 {
		panic("not pem encoded")
	}

	if block.Type != "RSA PRIVATE KEY" {
		panic("not rsa private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	return key
}

func MustParseECDSAPrivateKey(data string) *ecdsa.PrivateKey {
	block, _ := pem.Decode([]byte(data))
	if block == nil || block.Bytes == nil || len(block.Bytes) == 0 {
		panic("not pem encoded")
	}

	if block.Type != "EC PRIVATE KEY" {
		panic(fmt.Errorf("no ecdsa private key: %s", block.Type))
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	return key
}

func MustJSON(i any) string {
	if data, err := json.Marshal(i); err != nil {
		panic(err)
	} else {
		return string(data)
	}
}

func Test(t *testing.T) {
	config := &schema.OpenIDConnectConfiguration{
		IssuerCertificateChain: *MustParseX509CertificateChain(rsaChain),
		IssuerPrivateKey:       MustParseRSAPrivateKey(rsaKey),
		IssuerECDSA: []schema.ECPair{
			{
				CertificateChain: *MustParseX509CertificateChain(ecdsaChain),
				PrivateKey:       MustParseECDSAPrivateKey(ecdsaKey),
			},
		},
	}

	fmt.Println(config.IssuerPrivateKey.Size() * 8)

	strategy := NewKeyStrategy(config)

	fmt.Printf("%+v\n", strategy.alg)

	for kid, key := range strategy.keys {
		fmt.Printf("kid: %s, key: %T, alg: %s, use: %s\n", kid, key.key, key.alg, key.use)
		assert.NotNil(t, key.Key)

		fmt.Printf("%s\n", MustJSON(key.Public()))
	}

	headers := jwt.NewHeaders()

	headers.Extra = map[string]interface{}{
		JWTHeaderKeyIdentifier: "b51a4c",
	}

	token, sig, err := strategy.Generate(context.Background(), jwt.MapClaims{}, headers)

	assert.NoError(t, err)

	fmt.Println(token, sig)
}

import (
	"crypto"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestKeyManager_AddActiveJWK(t *testing.T) {
	manager := NewKeyManager()
	assert.Nil(t, manager.jwk)
	assert.Nil(t, manager.Strategy())

	j, err := manager.AddActiveJWK(schema.X509CertificateChain{}, mustParseRSAPrivateKey(exampleIssuerPrivateKey))
	require.NoError(t, err)
	require.NotNil(t, j)
	require.NotNil(t, manager.jwk)
	require.NotNil(t, manager.Strategy())

	thumbprint, err := j.JSONWebKey().Thumbprint(crypto.SHA1)
	assert.NoError(t, err)

	kid := strings.ToLower(fmt.Sprintf("%x", thumbprint)[:6])
	assert.Equal(t, manager.jwk.id, kid)
	assert.Equal(t, kid, j.JSONWebKey().KeyID)
	assert.Len(t, manager.jwks.Keys, 1)

	keys := manager.jwks.Key(kid)
	assert.Equal(t, keys[0].KeyID, kid)

	privKey, err := manager.GetActivePrivateKey()
	assert.NoError(t, err)
	assert.NotNil(t, privKey)

	webKey, err := manager.GetActiveJWK()
	assert.NoError(t, err)
	assert.NotNil(t, webKey)

	keySet := manager.GetKeySet()
	assert.NotNil(t, keySet)
	assert.Equal(t, kid, manager.GetActiveKeyID())
}
