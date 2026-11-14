// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package appauthenticator

import (
	"bytes"
	"context"
	"crypto/x509"
	"embed"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"sync"
	"time"
)

const (
	androidMaxChainLength = 10

	androidTagPurpose                  = 1
	androidTagAlgorithm                = 2
	androidTagKeySize                  = 3
	androidTagDigest                   = 5
	androidTagECCurve                  = 10
	androidTagNoAuthRequired           = 503
	androidTagUserAuthType             = 504
	androidTagAuthTimeout              = 505
	androidTagOrigin                   = 702
	androidTagAttestationApplicationID = 709

	androidPurposeSign           = 2
	androidAlgorithmEC           = 3
	androidECCurveP256           = 1
	androidKeySize               = 256
	androidDigestSHA256          = 4
	androidOriginGenerated       = 0
	androidUserAuthTypeBiometric = 2

	androidSecurityLevelSoftware           = 0
	androidSecurityLevelTrustedEnvironment = 1
	androidSecurityLevelStrongBox          = 2
)

var oidAndroidKeyDescription = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11129, 2, 1, 17}

//go:embed roots/google_*.pem
var googleRootsFS embed.FS

var googleRoots = sync.OnceValue(func() []*x509.Certificate {
	roots, err := parseGoogleRoots(googleRootsFS)
	if err != nil {
		panic(err)
	}

	return roots
})

// GoogleAttestationRoots returns the pinned Google hardware attestation root certificates.
func GoogleAttestationRoots() []*x509.Certificate {
	return slices.Clone(googleRoots())
}

func parseGoogleRoots(fsys fs.FS) (roots []*x509.Certificate, err error) {
	names, err := fs.Glob(fsys, "roots/google_*.pem")
	if err != nil {
		return nil, err
	}

	for _, name := range names {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, err
		}

		block, _ := pem.Decode(data)
		if block == nil || block.Type != "CERTIFICATE" {
			return nil, fmt.Errorf("error decoding root certificate '%s'", name)
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing root certificate '%s': %w", name, err)
		}

		roots = append(roots, cert)
	}

	return roots, nil
}

type keyDescription struct {
	AttestationVersion       int
	AttestationSecurityLevel asn1.Enumerated
	KeyMintVersion           int
	KeyMintSecurityLevel     asn1.Enumerated
	AttestationChallenge     []byte
	UniqueID                 []byte
	SoftwareEnforced         []asn1.RawValue
	HardwareEnforced         []asn1.RawValue
}

type attestationApplicationID struct {
	PackageInfos     []attestationPackageInfo `asn1:"set"`
	SignatureDigests [][]byte                 `asn1:"set"`
}

type attestationPackageInfo struct {
	PackageName []byte
	Version     int64
}

type authorizationList map[int][]byte

func parseAuthorizationList(elements []asn1.RawValue) (authorizationList, error) {
	list := make(authorizationList, len(elements))

	for _, element := range elements {
		if element.Class != asn1.ClassContextSpecific || !element.IsCompound {
			return nil, fmt.Errorf("authorization list element with tag %d is not explicitly tagged", element.Tag)
		}

		if _, ok := list[element.Tag]; ok {
			return nil, fmt.Errorf("authorization list contains tag %d more than once", element.Tag)
		}

		list[element.Tag] = element.Bytes
	}

	return list, nil
}

func (l authorizationList) has(tag int) bool {
	_, ok := l[tag]

	return ok
}

func (l authorizationList) integer(tag int) (value int64, ok bool, err error) {
	raw, ok := l[tag]
	if !ok {
		return 0, false, nil
	}

	if err = unmarshalStrict(raw, &value, ""); err != nil {
		return 0, true, fmt.Errorf("error parsing authorization list tag %d: %w", tag, err)
	}

	return value, true, nil
}

func (l authorizationList) integerSet(tag int) (values []int64, err error) {
	raw, ok := l[tag]
	if !ok {
		return nil, nil
	}

	if err = unmarshalStrict(raw, &values, "set"); err != nil {
		return nil, fmt.Errorf("error parsing authorization list tag %d: %w", tag, err)
	}

	return values, nil
}

func unmarshalStrict(data []byte, value any, params string) error {
	rest, err := asn1.UnmarshalWithParams(data, value, params)
	if err != nil {
		return err
	}

	if len(rest) != 0 {
		return errors.New("trailing data")
	}

	return nil
}

type androidChainResult struct {
	identities []int
	level      int
}

func (v *Verifier) verifyAndroidKey(ctx context.Context, in AttestationInput) (AttestationResult, error) {
	now := v.clock.Now()

	statuses := sync.OnceValues(func() (map[string]string, error) {
		return v.revocation.list(ctx, now)
	})

	approval, err := v.verifyAndroidChain(now, statuses, in.ApprovalChain, in.ApprovalSPKI, in.Challenge, true)
	if err != nil {
		return AttestationResult{}, fmt.Errorf("approval certificate chain: %w", err)
	}

	transport, err := v.verifyAndroidChain(now, statuses, in.TransportChain, in.TransportSPKI, in.Challenge, false)
	if err != nil {
		return AttestationResult{}, fmt.Errorf("transport certificate chain: %w", err)
	}

	for _, i := range approval.identities {
		if slices.Contains(transport.identities, i) {
			return AttestationResult{
				Format:   AttestationFormatAndroidKey,
				Identity: v.config.Android[i].PackageName,
				Detail:   androidSecurityLevelName(min(approval.level, transport.level)),
				Attested: true,
			}, nil
		}
	}

	return AttestationResult{}, errors.New("certificate chains match different android identities")
}

func (v *Verifier) verifyAndroidChain(now time.Time, statuses func() (map[string]string, error), encoded []string, spki []byte, challenge string, approval bool) (result androidChainResult, err error) {
	certs, err := decodeAndroidChain(encoded)
	if err != nil {
		return result, err
	}

	if err = v.verifyAndroidChainTrust(now, certs); err != nil {
		return result, err
	}

	if err = checkAndroidRevocation(statuses, certs); err != nil {
		return result, err
	}

	if !bytes.Equal(certs[0].RawSubjectPublicKeyInfo, spki) {
		return result, errors.New("leaf public key does not match the submitted public key")
	}

	kd, err := parseAndroidKeyDescription(certs)
	if err != nil {
		return result, err
	}

	if !bytes.Equal(kd.AttestationChallenge, []byte(challenge)) {
		return result, errors.New("attestation challenge does not match")
	}

	if result.level, err = androidSecurityLevel(kd); err != nil {
		return result, err
	}

	hardware, err := parseAuthorizationList(kd.HardwareEnforced)
	if err != nil {
		return result, fmt.Errorf("hardware enforced %w", err)
	}

	software, err := parseAuthorizationList(kd.SoftwareEnforced)
	if err != nil {
		return result, fmt.Errorf("software enforced %w", err)
	}

	if err = checkAndroidHardwareEnforced(hardware); err != nil {
		return result, err
	}

	if result.identities, err = v.matchAndroidIdentities(software); err != nil {
		return result, err
	}

	if approval {
		err = checkAndroidApprovalAuthorization(hardware, software)
	} else if !hardware.has(androidTagNoAuthRequired) {
		err = errors.New("no auth required is absent")
	}

	return result, err
}

func decodeAndroidChain(encoded []string) ([]*x509.Certificate, error) {
	switch {
	case len(encoded) == 0:
		return nil, errors.New("chain is empty")
	case len(encoded) == 1:
		return nil, errors.New("chain is too short")
	case len(encoded) > androidMaxChainLength:
		return nil, errors.New("chain is too long")
	}

	certs := make([]*x509.Certificate, len(encoded))

	for i, value := range encoded {
		der, err := base64.RawURLEncoding.DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("certificate %d is not valid base64url", i)
		}

		if certs[i], err = x509.ParseCertificate(der); err != nil {
			return nil, fmt.Errorf("error parsing certificate %d: %w", i, err)
		}
	}

	return certs, nil
}

func (v *Verifier) verifyAndroidChainTrust(now time.Time, certs []*x509.Certificate) error {
	for i := 0; i < len(certs)-1; i++ {
		if err := certs[i+1].CheckSignature(certs[i].SignatureAlgorithm, certs[i].RawTBSCertificate, certs[i].Signature); err != nil {
			return fmt.Errorf("certificate %d signature is not valid: %w", i, err)
		}
	}

	for i, cert := range certs[:len(certs)-1] {
		if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
			return fmt.Errorf("certificate %d is not valid at the current time", i)
		}
	}

	last := certs[len(certs)-1]

	for _, root := range v.androidRoots {
		if bytes.Equal(root.RawSubjectPublicKeyInfo, last.RawSubjectPublicKeyInfo) {
			return nil
		}
	}

	return errors.New("chain does not end in a trusted root")
}

func parseAndroidKeyDescription(certs []*x509.Certificate) (kd keyDescription, err error) {
	for i, cert := range certs[1:] {
		if findExtension(cert, oidAndroidKeyDescription) != nil {
			return kd, fmt.Errorf("certificate %d contains the key description extension", i+1)
		}

		if !cert.BasicConstraintsValid || !cert.IsCA {
			return kd, fmt.Errorf("certificate %d is not a certificate authority", i+1)
		}

		if cert.KeyUsage&x509.KeyUsageCertSign == 0 {
			return kd, fmt.Errorf("certificate %d key usage does not include certificate signing", i+1)
		}
	}

	ext := findExtension(certs[0], oidAndroidKeyDescription)
	if ext == nil {
		return kd, errors.New("leaf certificate does not contain the key description extension")
	}

	if err = unmarshalStrict(ext, &kd, ""); err != nil {
		return kd, fmt.Errorf("error parsing key description: %w", err)
	}

	return kd, nil
}

func findExtension(cert *x509.Certificate, oid asn1.ObjectIdentifier) []byte {
	for _, ext := range cert.Extensions {
		if ext.Id.Equal(oid) {
			return ext.Value
		}
	}

	return nil
}

func androidSecurityLevel(kd keyDescription) (int, error) {
	for _, level := range []struct {
		name  string
		value asn1.Enumerated
	}{
		{"attestation", kd.AttestationSecurityLevel},
		{"keymint", kd.KeyMintSecurityLevel},
	} {
		if level.value != androidSecurityLevelTrustedEnvironment && level.value != androidSecurityLevelStrongBox {
			return 0, fmt.Errorf("%s security level '%s' is not hardware backed", level.name, androidSecurityLevelName(int(level.value)))
		}
	}

	return int(min(kd.AttestationSecurityLevel, kd.KeyMintSecurityLevel)), nil
}

func androidSecurityLevelName(level int) string {
	switch level {
	case androidSecurityLevelSoftware:
		return "Software"
	case androidSecurityLevelTrustedEnvironment:
		return "TrustedEnvironment"
	case androidSecurityLevelStrongBox:
		return "StrongBox"
	default:
		return fmt.Sprintf("Unknown(%d)", level)
	}
}

func checkAndroidHardwareEnforced(hardware authorizationList) error {
	for _, set := range []struct {
		tag   int
		value int64
		err   string
	}{
		{androidTagPurpose, androidPurposeSign, "purpose does not contain sign"},
		{androidTagDigest, androidDigestSHA256, "digest does not contain SHA-256"},
	} {
		values, err := hardware.integerSet(set.tag)
		if err != nil {
			return err
		}

		if !slices.Contains(values, set.value) {
			return errors.New(set.err)
		}
	}

	for _, value := range []struct {
		tag   int
		name  string
		value int64
		want  string
	}{
		{androidTagAlgorithm, "algorithm", androidAlgorithmEC, "EC"},
		{androidTagECCurve, "ec curve", androidECCurveP256, "P-256"},
		{androidTagKeySize, "key size", androidKeySize, "256"},
		{androidTagOrigin, "origin", androidOriginGenerated, "generated"},
	} {
		actual, ok, err := hardware.integer(value.tag)
		if err != nil {
			return err
		}

		if !ok {
			return fmt.Errorf("%s is absent", value.name)
		}

		if actual != value.value {
			return fmt.Errorf("%s %d is not %s", value.name, actual, value.want)
		}
	}

	return nil
}

func checkAndroidApprovalAuthorization(hardware, software authorizationList) error {
	if hardware.has(androidTagNoAuthRequired) || software.has(androidTagNoAuthRequired) {
		return errors.New("no auth required is present")
	}

	userAuthType, ok, err := hardware.integer(androidTagUserAuthType)
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("user auth type is absent")
	}

	if userAuthType != androidUserAuthTypeBiometric {
		return fmt.Errorf("user auth type %d is not biometric only", userAuthType)
	}

	for _, list := range []authorizationList{hardware, software} {
		timeout, _, err := list.integer(androidTagAuthTimeout)
		if err != nil {
			return err
		}

		if timeout != 0 {
			return fmt.Errorf("auth timeout %d is not zero", timeout)
		}
	}

	return nil
}

func parseAttestationApplicationID(software authorizationList) (appID attestationApplicationID, err error) {
	raw, ok := software[androidTagAttestationApplicationID]
	if !ok {
		return appID, errors.New("attestation application id is absent")
	}

	var encoded []byte

	if err = unmarshalStrict(raw, &encoded, ""); err != nil {
		return appID, fmt.Errorf("error parsing attestation application id: %w", err)
	}

	if err = unmarshalStrict(encoded, &appID, ""); err != nil {
		return appID, fmt.Errorf("error parsing attestation application id: %w", err)
	}

	return appID, nil
}

func (v *Verifier) matchAndroidIdentities(software authorizationList) (identities []int, err error) {
	appID, err := parseAttestationApplicationID(software)
	if err != nil {
		return nil, err
	}

	for i, identity := range v.config.Android {
		if matchAndroidIdentity(identity.PackageName, identity.SigningCertificateDigests, appID) {
			identities = append(identities, i)
		}
	}

	if len(identities) == 0 {
		return nil, errors.New("attestation application id does not match any configured android identity")
	}

	return identities, nil
}

func matchAndroidIdentity(packageName string, digests []string, appID attestationApplicationID) bool {
	if len(appID.SignatureDigests) == 0 {
		return false
	}

	if !slices.ContainsFunc(appID.PackageInfos, func(info attestationPackageInfo) bool {
		return string(info.PackageName) == packageName
	}) {
		return false
	}

	for _, digest := range appID.SignatureDigests {
		if !slices.Contains(digests, hex.EncodeToString(digest)) {
			return false
		}
	}

	return true
}
