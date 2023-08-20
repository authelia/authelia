package validator

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/ory/fosite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateIdentityProviders validates and updates the IdentityProviders configuration.
func ValidateIdentityProviders(config *schema.IdentityProviders, val *schema.StructValidator) {
	validateOIDC(config.OIDC, val)
}

func validateOIDC(config *schema.OpenIDConnect, val *schema.StructValidator) {
	if config == nil {
		return
	}

	setOIDCDefaults(config)

	validateOIDCIssuer(config, val)
	validateOIDCAuthorizationPolicies(config, val)
	validateOIDCLifespans(config, val)

	sort.Sort(oidc.SortedSigningAlgs(config.Discovery.ResponseObjectSigningAlgs))

	switch {
	case config.MinimumParameterEntropy == -1:
		val.PushWarning(fmt.Errorf(errFmtOIDCProviderInsecureDisabledParameterEntropy))
	case config.MinimumParameterEntropy <= 0:
		config.MinimumParameterEntropy = fosite.MinParameterEntropy
	case config.MinimumParameterEntropy < fosite.MinParameterEntropy:
		val.PushWarning(fmt.Errorf(errFmtOIDCProviderInsecureParameterEntropy, fosite.MinParameterEntropy, config.MinimumParameterEntropy))
	}

	switch config.EnforcePKCE {
	case "always", "never", "public_clients_only":
		break
	default:
		val.Push(fmt.Errorf(errFmtOIDCProviderEnforcePKCEInvalidValue, config.EnforcePKCE))
	}

	validateOIDCOptionsCORS(config, val)

	if len(config.Clients) == 0 {
		val.Push(fmt.Errorf(errFmtOIDCProviderNoClientsConfigured))
	} else {
		validateOIDCClients(config, val)
	}
}

func validateOIDCAuthorizationPolicies(config *schema.OpenIDConnect, val *schema.StructValidator) {
	config.Discovery.AuthorizationPolicies = []string{policyOneFactor, policyTwoFactor}

	for name, policy := range config.AuthorizationPolicies {
		switch name {
		case "":
			val.Push(fmt.Errorf(errFmtOIDCPolicyInvalidName))
		case policyOneFactor, policyTwoFactor, policyDeny:
			val.Push(fmt.Errorf(errFmtOIDCPolicyInvalidNameStandard, name, "name", strJoinAnd([]string{policyOneFactor, policyTwoFactor, policyDeny}), name))
		}

		switch policy.DefaultPolicy {
		case "":
			policy.DefaultPolicy = schema.DefaultOpenIDConnectPolicyConfiguration.DefaultPolicy
		case policyOneFactor, policyTwoFactor, policyDeny:
			break
		default:
			val.Push(fmt.Errorf(errFmtOIDCPolicyInvalidDefaultPolicy, name, strJoinAnd([]string{policyOneFactor, policyTwoFactor, policyDeny}), policy.DefaultPolicy))
		}

		if len(policy.Rules) == 0 {
			val.Push(fmt.Errorf(errFmtOIDCPolicyMissingOption, name, "rules"))
		}

		for i, rule := range policy.Rules {
			switch rule.Policy {
			case "":
				policy.Rules[i].Policy = schema.DefaultOpenIDConnectPolicyConfiguration.DefaultPolicy
			case policyOneFactor, policyTwoFactor, policyDeny:
				break
			default:
				val.Push(fmt.Errorf(errFmtOIDCPolicyRuleInvalidPolicy, name, i+1, strJoinAnd([]string{policyOneFactor, policyTwoFactor, policyDeny}), rule.Policy))
			}

			if len(rule.Subjects) == 0 {
				val.Push(fmt.Errorf(errFmtOIDCPolicyRuleMissingOption, name, i+1, "subject"))
			}
		}

		config.AuthorizationPolicies[name] = policy

		config.Discovery.AuthorizationPolicies = append(config.Discovery.AuthorizationPolicies, name)
	}
}

func validateOIDCLifespans(config *schema.OpenIDConnect, _ *schema.StructValidator) {
	for name := range config.Lifespans.Custom {
		config.Discovery.Lifespans = append(config.Discovery.Lifespans, name)
	}
}

func validateOIDCIssuer(config *schema.OpenIDConnect, val *schema.StructValidator) {
	switch {
	case config.IssuerPrivateKey != nil:
		validateOIDCIssuerPrivateKey(config)

		fallthrough
	case len(config.IssuerPrivateKeys) != 0:
		validateOIDCIssuerPrivateKeys(config, val)
	default:
		val.Push(fmt.Errorf(errFmtOIDCProviderNoPrivateKey))
	}
}

func validateOIDCIssuerPrivateKey(config *schema.OpenIDConnect) {
	config.IssuerPrivateKeys = append([]schema.JWK{{
		Algorithm:        oidc.SigningAlgRSAUsingSHA256,
		Use:              oidc.KeyUseSignature,
		Key:              config.IssuerPrivateKey,
		CertificateChain: config.IssuerCertificateChain,
	}}, config.IssuerPrivateKeys...)
}

func validateOIDCIssuerPrivateKeys(config *schema.OpenIDConnect, val *schema.StructValidator) {
	var (
		props *JWKProperties
		err   error
	)

	config.Discovery.ResponseObjectSigningKeyIDs = make([]string, len(config.IssuerPrivateKeys))
	config.Discovery.DefaultKeyIDs = map[string]string{}

	for i := 0; i < len(config.IssuerPrivateKeys); i++ {
		if key, ok := config.IssuerPrivateKeys[i].Key.(*rsa.PrivateKey); ok && key.PublicKey.N == nil {
			val.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysInvalid, i+1))

			continue
		}

		switch n := len(config.IssuerPrivateKeys[i].KeyID); {
		case n == 0:
			if config.IssuerPrivateKeys[i].KeyID, err = jwkCalculateThumbprint(config.IssuerPrivateKeys[i].Key); err != nil {
				val.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysCalcThumbprint, i+1, err))

				continue
			}
		case n > 100:
			val.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysKeyIDLength, i+1, config.IssuerPrivateKeys[i].KeyID))
		}

		if config.IssuerPrivateKeys[i].KeyID != "" && utils.IsStringInSlice(config.IssuerPrivateKeys[i].KeyID, config.Discovery.ResponseObjectSigningKeyIDs) {
			val.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysAttributeNotUnique, i+1, config.IssuerPrivateKeys[i].KeyID, attrOIDCKeyID))
		}

		config.Discovery.ResponseObjectSigningKeyIDs[i] = config.IssuerPrivateKeys[i].KeyID

		if !reOpenIDConnectKID.MatchString(config.IssuerPrivateKeys[i].KeyID) {
			val.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysKeyIDNotValid, i+1, config.IssuerPrivateKeys[i].KeyID))
		}

		if props, err = schemaJWKGetProperties(config.IssuerPrivateKeys[i]); err != nil {
			val.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysProperties, i+1, config.IssuerPrivateKeys[i].KeyID, err))

			continue
		}

		validateOIDCIssuerPrivateKeysUseAlg(i, props, config, val)
		validateOIDCIssuerPrivateKeyPair(i, config, val)
	}

	if len(config.Discovery.ResponseObjectSigningAlgs) != 0 && !utils.IsStringInSlice(oidc.SigningAlgRSAUsingSHA256, config.Discovery.ResponseObjectSigningAlgs) {
		val.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysNoRS256, oidc.SigningAlgRSAUsingSHA256, strJoinAnd(config.Discovery.ResponseObjectSigningAlgs)))
	}
}

func validateOIDCIssuerPrivateKeysUseAlg(i int, props *JWKProperties, config *schema.OpenIDConnect, val *schema.StructValidator) {
	switch config.IssuerPrivateKeys[i].Use {
	case "":
		config.IssuerPrivateKeys[i].Use = props.Use
	case oidc.KeyUseSignature:
		break
	default:
		val.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysInvalidOptionOneOf, i+1, config.IssuerPrivateKeys[i].KeyID, attrOIDCKeyUse, strJoinOr([]string{oidc.KeyUseSignature}), config.IssuerPrivateKeys[i].Use))
	}

	switch {
	case config.IssuerPrivateKeys[i].Algorithm == "":
		config.IssuerPrivateKeys[i].Algorithm = props.Algorithm

		fallthrough
	case utils.IsStringInSlice(config.IssuerPrivateKeys[i].Algorithm, validOIDCIssuerJWKSigningAlgs):
		if config.IssuerPrivateKeys[i].KeyID != "" && config.IssuerPrivateKeys[i].Algorithm != "" {
			if _, ok := config.Discovery.DefaultKeyIDs[config.IssuerPrivateKeys[i].Algorithm]; !ok {
				config.Discovery.DefaultKeyIDs[config.IssuerPrivateKeys[i].Algorithm] = config.IssuerPrivateKeys[i].KeyID
			}
		}
	default:
		val.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysInvalidOptionOneOf, i+1, config.IssuerPrivateKeys[i].KeyID, attrOIDCAlgorithm, strJoinOr(validOIDCIssuerJWKSigningAlgs), config.IssuerPrivateKeys[i].Algorithm))
	}

	if config.IssuerPrivateKeys[i].Algorithm != "" {
		if !utils.IsStringInSlice(config.IssuerPrivateKeys[i].Algorithm, config.Discovery.ResponseObjectSigningAlgs) {
			config.Discovery.ResponseObjectSigningAlgs = append(config.Discovery.ResponseObjectSigningAlgs, config.IssuerPrivateKeys[i].Algorithm)
		}
	}
}

func validateOIDCIssuerPrivateKeyPair(i int, config *schema.OpenIDConnect, val *schema.StructValidator) {
	var (
		checkEqualKey bool
		err           error
	)

	switch key := config.IssuerPrivateKeys[i].Key.(type) {
	case *rsa.PrivateKey:
		checkEqualKey = true

		if key.Size() < 256 {
			checkEqualKey = false

			val.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysRSAKeyLessThan2048Bits, i+1, config.IssuerPrivateKeys[i].KeyID, key.Size()*8))
		}
	case *ecdsa.PrivateKey:
		checkEqualKey = true
	default:
		val.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysKeyNotRSAOrECDSA, i+1, config.IssuerPrivateKeys[i].KeyID, key))
	}

	if config.IssuerPrivateKeys[i].CertificateChain.HasCertificates() {
		if checkEqualKey && !config.IssuerPrivateKeys[i].CertificateChain.EqualKey(config.IssuerPrivateKeys[i].Key) {
			val.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysKeyCertificateMismatch, i+1, config.IssuerPrivateKeys[i].KeyID))
		}

		if err = config.IssuerPrivateKeys[i].CertificateChain.Validate(); err != nil {
			val.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysCertificateChainInvalid, i+1, config.IssuerPrivateKeys[i].KeyID, err))
		}
	}
}

func setOIDCDefaults(config *schema.OpenIDConnect) {
	if config.Lifespans.AccessToken == durationZero {
		config.Lifespans.AccessToken = schema.DefaultOpenIDConnectConfiguration.Lifespans.AccessToken
	}

	if config.Lifespans.AuthorizeCode == durationZero {
		config.Lifespans.AuthorizeCode = schema.DefaultOpenIDConnectConfiguration.Lifespans.AuthorizeCode
	}

	if config.Lifespans.IDToken == durationZero {
		config.Lifespans.IDToken = schema.DefaultOpenIDConnectConfiguration.Lifespans.IDToken
	}

	if config.Lifespans.RefreshToken == durationZero {
		config.Lifespans.RefreshToken = schema.DefaultOpenIDConnectConfiguration.Lifespans.RefreshToken
	}

	if config.EnforcePKCE == "" {
		config.EnforcePKCE = schema.DefaultOpenIDConnectConfiguration.EnforcePKCE
	}
}

func validateOIDCOptionsCORS(config *schema.OpenIDConnect, validator *schema.StructValidator) {
	validateOIDCOptionsCORSAllowedOrigins(config, validator)

	if config.CORS.AllowedOriginsFromClientRedirectURIs {
		validateOIDCOptionsCORSAllowedOriginsFromClientRedirectURIs(config)
	}

	validateOIDCOptionsCORSEndpoints(config, validator)
}

func validateOIDCOptionsCORSAllowedOrigins(config *schema.OpenIDConnect, val *schema.StructValidator) {
	for _, origin := range config.CORS.AllowedOrigins {
		if origin.String() == "*" {
			if len(config.CORS.AllowedOrigins) != 1 {
				val.Push(fmt.Errorf(errFmtOIDCCORSInvalidOriginWildcard))
			}

			if config.CORS.AllowedOriginsFromClientRedirectURIs {
				val.Push(fmt.Errorf(errFmtOIDCCORSInvalidOriginWildcardWithClients))
			}

			continue
		}

		if origin.Path != "" {
			val.Push(fmt.Errorf(errFmtOIDCCORSInvalidOrigin, origin.String(), "path"))
		}

		if origin.RawQuery != "" {
			val.Push(fmt.Errorf(errFmtOIDCCORSInvalidOrigin, origin.String(), "query string"))
		}
	}
}

func validateOIDCOptionsCORSAllowedOriginsFromClientRedirectURIs(config *schema.OpenIDConnect) {
	for _, client := range config.Clients {
		for _, redirectURI := range client.RedirectURIs {
			uri, err := url.ParseRequestURI(redirectURI)
			if err != nil || (uri.Scheme != schemeHTTP && uri.Scheme != schemeHTTPS) || uri.Hostname() == "localhost" {
				continue
			}

			origin := utils.OriginFromURL(uri)

			if !utils.IsURLInSlice(*origin, config.CORS.AllowedOrigins) {
				config.CORS.AllowedOrigins = append(config.CORS.AllowedOrigins, *origin)
			}
		}
	}
}

func validateOIDCOptionsCORSEndpoints(config *schema.OpenIDConnect, val *schema.StructValidator) {
	for _, endpoint := range config.CORS.Endpoints {
		if !utils.IsStringInSlice(endpoint, validOIDCCORSEndpoints) {
			val.Push(fmt.Errorf(errFmtOIDCCORSInvalidEndpoint, endpoint, strJoinOr(validOIDCCORSEndpoints)))
		}
	}
}

func validateOIDCClients(config *schema.OpenIDConnect, val *schema.StructValidator) {
	var (
		errDeprecated bool

		clientIDs, duplicateClientIDs, blankClientIDs []string
	)

	errDeprecatedFunc := func() { errDeprecated = true }

	for c, client := range config.Clients {
		if client.ID == "" {
			blankClientIDs = append(blankClientIDs, "#"+strconv.Itoa(c+1))
		} else {
			if client.Description == "" {
				config.Clients[c].Description = client.ID
			}

			if id := strings.ToLower(client.ID); utils.IsStringInSlice(id, clientIDs) {
				if !utils.IsStringInSlice(id, duplicateClientIDs) {
					duplicateClientIDs = append(duplicateClientIDs, id)
				}
			} else {
				clientIDs = append(clientIDs, id)
			}
		}

		validateOIDCClient(c, config, val, errDeprecatedFunc)
	}

	if errDeprecated {
		val.PushWarning(fmt.Errorf(errFmtOIDCClientsDeprecated))
	}

	if len(blankClientIDs) != 0 {
		val.Push(fmt.Errorf(errFmtOIDCClientsWithEmptyID, buildJoinedString(", ", "or", "", blankClientIDs)))
	}

	if len(duplicateClientIDs) != 0 {
		val.Push(fmt.Errorf(errFmtOIDCClientsDuplicateID, strJoinOr(duplicateClientIDs)))
	}
}

func validateOIDCClient(c int, config *schema.OpenIDConnect, val *schema.StructValidator, errDeprecatedFunc func()) {
	if config.Clients[c].Public {
		if config.Clients[c].Secret != nil {
			val.Push(fmt.Errorf(errFmtOIDCClientPublicInvalidSecret, config.Clients[c].ID))
		}
	} else {
		if config.Clients[c].Secret == nil {
			val.Push(fmt.Errorf(errFmtOIDCClientInvalidSecret, config.Clients[c].ID))
		} else {
			switch {
			case config.Clients[c].Secret.IsPlainText() && config.Clients[c].TokenEndpointAuthMethod != oidc.ClientAuthMethodClientSecretJWT:
				val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidSecretPlainText, config.Clients[c].ID))
			case !config.Clients[c].Secret.IsPlainText() && config.Clients[c].TokenEndpointAuthMethod == oidc.ClientAuthMethodClientSecretJWT:
				val.Push(fmt.Errorf(errFmtOIDCClientInvalidSecretNotPlainText, config.Clients[c].ID))
			}
		}
	}

	switch {
	case config.Clients[c].AuthorizationPolicy == "":
		config.Clients[c].AuthorizationPolicy = schema.DefaultOpenIDConnectClientConfiguration.AuthorizationPolicy
	case utils.IsStringInSlice(config.Clients[c].AuthorizationPolicy, config.Discovery.AuthorizationPolicies):
		break
	default:
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidValue, config.Clients[c].ID, "authorization_policy", strJoinOr(config.Discovery.AuthorizationPolicies), config.Clients[c].AuthorizationPolicy))
	}

	switch {
	case config.Clients[c].Lifespan == "", utils.IsStringInSlice(config.Clients[c].Lifespan, config.Discovery.Lifespans):
		break
	default:
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidValue, config.Clients[c].ID, "lifespan", strJoinOr(config.Discovery.Lifespans), config.Clients[c].Lifespan))
	}

	switch config.Clients[c].PKCEChallengeMethod {
	case "", oidc.PKCEChallengeMethodPlain, oidc.PKCEChallengeMethodSHA256:
		break
	default:
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidValue, config.Clients[c].ID, attrOIDCPKCEChallengeMethod, strJoinOr([]string{oidc.PKCEChallengeMethodPlain, oidc.PKCEChallengeMethodSHA256}), config.Clients[c].PKCEChallengeMethod))
	}

	validateOIDCClientConsentMode(c, config, val)

	validateOIDCClientScopes(c, config, val, errDeprecatedFunc)
	validateOIDCClientResponseTypes(c, config, val, errDeprecatedFunc)
	validateOIDCClientResponseModes(c, config, val, errDeprecatedFunc)
	validateOIDCClientGrantTypes(c, config, val, errDeprecatedFunc)
	validateOIDCClientRedirectURIs(c, config, val, errDeprecatedFunc)

	validateOIDDClientSigningAlgs(c, config, val)

	validateOIDCClientSectorIdentifier(c, config, val)

	validateOIDCClientPublicKeys(c, config, val)
	validateOIDCClientTokenEndpointAuth(c, config, val)
}

func validateOIDCClientPublicKeys(c int, config *schema.OpenIDConnect, val *schema.StructValidator) {
	switch {
	case config.Clients[c].PublicKeys.URI != nil && len(config.Clients[c].PublicKeys.Values) != 0:
		val.Push(fmt.Errorf(errFmtOIDCClientPublicKeysBothURIAndValuesConfigured, config.Clients[c].ID))
	case config.Clients[c].PublicKeys.URI != nil:
		if config.Clients[c].PublicKeys.URI.Scheme != schemeHTTPS {
			val.Push(fmt.Errorf(errFmtOIDCClientPublicKeysURIInvalidScheme, config.Clients[c].ID, config.Clients[c].PublicKeys.URI.Scheme))
		}
	case len(config.Clients[c].PublicKeys.Values) != 0:
		validateOIDCClientJSONWebKeysList(c, config, val)
	}
}

func validateOIDCClientJSONWebKeysList(c int, config *schema.OpenIDConnect, val *schema.StructValidator) {
	var (
		props *JWKProperties
		err   error
	)

	for i := 0; i < len(config.Clients[c].PublicKeys.Values); i++ {
		if config.Clients[c].PublicKeys.Values[i].KeyID == "" {
			val.Push(fmt.Errorf(errFmtOIDCClientPublicKeysInvalidOptionMissingOneOf, config.Clients[c].ID, i+1, attrOIDCKeyID))
		}

		if props, err = schemaJWKGetProperties(config.Clients[c].PublicKeys.Values[i]); err != nil {
			val.Push(fmt.Errorf(errFmtOIDCClientPublicKeysProperties, config.Clients[c].ID, i+1, config.Clients[c].PublicKeys.Values[i].KeyID, err))

			continue
		}

		validateOIDCClientJSONWebKeysListKeyUseAlg(c, i, props, config, val)

		var checkEqualKey bool

		switch key := config.Clients[c].PublicKeys.Values[i].Key.(type) {
		case nil:
			val.Push(fmt.Errorf(errFmtOIDCClientPublicKeysInvalidOptionMissingOneOf, config.Clients[c].ID, i+1, attrOIDCKey))
		case *rsa.PublicKey:
			checkEqualKey = true

			if key.N == nil {
				checkEqualKey = false

				val.Push(fmt.Errorf(errFmtOIDCClientPublicKeysKeyMalformed, config.Clients[c].ID, i+1))
			} else if key.Size() < 256 {
				checkEqualKey = false

				val.Push(fmt.Errorf(errFmtOIDCClientPublicKeysRSAKeyLessThan2048Bits, config.Clients[c].ID, i+1, config.Clients[c].PublicKeys.Values[i].KeyID, key.Size()*8))
			}
		case *ecdsa.PublicKey:
			checkEqualKey = true
		default:
			val.Push(fmt.Errorf(errFmtOIDCClientPublicKeysKeyNotRSAOrECDSA, config.Clients[c].ID, i+1, config.Clients[c].PublicKeys.Values[i].KeyID, key))
		}

		if config.Clients[c].PublicKeys.Values[i].CertificateChain.HasCertificates() {
			if checkEqualKey && !config.Clients[c].PublicKeys.Values[i].CertificateChain.EqualKey(config.Clients[c].PublicKeys.Values[i].Key) {
				val.Push(fmt.Errorf(errFmtOIDCClientPublicKeysCertificateChainKeyMismatch, config.Clients[c].ID, i+1, config.Clients[c].PublicKeys.Values[i].KeyID))
			}

			if err = config.Clients[c].PublicKeys.Values[i].CertificateChain.Validate(); err != nil {
				val.Push(fmt.Errorf(errFmtOIDCClientPublicKeysCertificateChainInvalid, config.Clients[c].ID, i+1, config.Clients[c].PublicKeys.Values[i].KeyID, err))
			}
		}
	}

	if config.Clients[c].RequestObjectSigningAlg != "" && !utils.IsStringInSlice(config.Clients[c].RequestObjectSigningAlg, config.Clients[c].Discovery.RequestObjectSigningAlgs) {
		val.Push(fmt.Errorf(errFmtOIDCClientPublicKeysROSAMissingAlgorithm, config.Clients[c].ID, strJoinOr(config.Clients[c].Discovery.RequestObjectSigningAlgs)))
	}
}

func validateOIDCClientJSONWebKeysListKeyUseAlg(c, i int, props *JWKProperties, config *schema.OpenIDConnect, val *schema.StructValidator) {
	switch config.Clients[c].PublicKeys.Values[i].Use {
	case "":
		config.Clients[c].PublicKeys.Values[i].Use = props.Use
	case oidc.KeyUseSignature:
		break
	default:
		val.Push(fmt.Errorf(errFmtOIDCClientPublicKeysInvalidOptionOneOf, config.Clients[c].ID, i+1, config.Clients[c].PublicKeys.Values[i].KeyID, attrOIDCKeyUse, strJoinOr([]string{oidc.KeyUseSignature}), config.Clients[c].PublicKeys.Values[i].Use))
	}

	switch {
	case config.Clients[c].PublicKeys.Values[i].Algorithm == "":
		config.Clients[c].PublicKeys.Values[i].Algorithm = props.Algorithm
	case utils.IsStringInSlice(config.Clients[c].PublicKeys.Values[i].Algorithm, validOIDCIssuerJWKSigningAlgs):
		break
	default:
		val.Push(fmt.Errorf(errFmtOIDCClientPublicKeysInvalidOptionOneOf, config.Clients[c].ID, i+1, config.Clients[c].PublicKeys.Values[i].KeyID, attrOIDCAlgorithm, strJoinOr(validOIDCIssuerJWKSigningAlgs), config.Clients[c].PublicKeys.Values[i].Algorithm))
	}

	if config.Clients[c].PublicKeys.Values[i].Algorithm != "" {
		if !utils.IsStringInSlice(config.Clients[c].PublicKeys.Values[i].Algorithm, config.Discovery.RequestObjectSigningAlgs) {
			config.Discovery.RequestObjectSigningAlgs = append(config.Discovery.RequestObjectSigningAlgs, config.Clients[c].PublicKeys.Values[i].Algorithm)
		}

		if !utils.IsStringInSlice(config.Clients[c].PublicKeys.Values[i].Algorithm, config.Clients[c].Discovery.RequestObjectSigningAlgs) {
			config.Clients[c].Discovery.RequestObjectSigningAlgs = append(config.Clients[c].Discovery.RequestObjectSigningAlgs, config.Clients[c].PublicKeys.Values[i].Algorithm)
		}
	}
}

func validateOIDCClientSectorIdentifier(c int, config *schema.OpenIDConnect, val *schema.StructValidator) {
	if config.Clients[c].SectorIdentifier.String() != "" {
		if utils.IsURLHostComponent(config.Clients[c].SectorIdentifier) || utils.IsURLHostComponentWithPort(config.Clients[c].SectorIdentifier) {
			return
		}

		if config.Clients[c].SectorIdentifier.Scheme != "" {
			val.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifier, config.Clients[c].ID, config.Clients[c].SectorIdentifier.String(), config.Clients[c].SectorIdentifier.Host, "scheme", config.Clients[c].SectorIdentifier.Scheme))

			if config.Clients[c].SectorIdentifier.Path != "" {
				val.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifier, config.Clients[c].ID, config.Clients[c].SectorIdentifier.String(), config.Clients[c].SectorIdentifier.Host, "path", config.Clients[c].SectorIdentifier.Path))
			}

			if config.Clients[c].SectorIdentifier.RawQuery != "" {
				val.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifier, config.Clients[c].ID, config.Clients[c].SectorIdentifier.String(), config.Clients[c].SectorIdentifier.Host, "query", config.Clients[c].SectorIdentifier.RawQuery))
			}

			if config.Clients[c].SectorIdentifier.Fragment != "" {
				val.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifier, config.Clients[c].ID, config.Clients[c].SectorIdentifier.String(), config.Clients[c].SectorIdentifier.Host, "fragment", config.Clients[c].SectorIdentifier.Fragment))
			}

			if config.Clients[c].SectorIdentifier.User != nil {
				if config.Clients[c].SectorIdentifier.User.Username() != "" {
					val.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifier, config.Clients[c].ID, config.Clients[c].SectorIdentifier.String(), config.Clients[c].SectorIdentifier.Host, "username", config.Clients[c].SectorIdentifier.User.Username()))
				}

				if _, set := config.Clients[c].SectorIdentifier.User.Password(); set {
					val.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifierWithoutValue, config.Clients[c].ID, config.Clients[c].SectorIdentifier.String(), config.Clients[c].SectorIdentifier.Host, "password"))
				}
			}
		} else if config.Clients[c].SectorIdentifier.Host == "" {
			val.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifierHost, config.Clients[c].ID, config.Clients[c].SectorIdentifier.String()))
		}
	}
}

func validateOIDCClientConsentMode(c int, config *schema.OpenIDConnect, val *schema.StructValidator) {
	switch {
	case utils.IsStringInSlice(config.Clients[c].ConsentMode, []string{"", auto}):
		if config.Clients[c].ConsentPreConfiguredDuration != nil {
			config.Clients[c].ConsentMode = oidc.ClientConsentModePreConfigured.String()
		} else {
			config.Clients[c].ConsentMode = oidc.ClientConsentModeExplicit.String()
		}
	case utils.IsStringInSlice(config.Clients[c].ConsentMode, validOIDCClientConsentModes):
		break
	default:
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidConsentMode, config.Clients[c].ID, strJoinOr(append(validOIDCClientConsentModes, auto)), config.Clients[c].ConsentMode))
	}

	if config.Clients[c].ConsentMode == oidc.ClientConsentModePreConfigured.String() && config.Clients[c].ConsentPreConfiguredDuration == nil {
		config.Clients[c].ConsentPreConfiguredDuration = schema.DefaultOpenIDConnectClientConfiguration.ConsentPreConfiguredDuration
	}
}

func validateOIDCClientScopes(c int, config *schema.OpenIDConnect, val *schema.StructValidator, errDeprecatedFunc func()) {
	if len(config.Clients[c].Scopes) == 0 {
		config.Clients[c].Scopes = schema.DefaultOpenIDConnectClientConfiguration.Scopes
	}

	invalid, duplicates := validateList(config.Clients[c].Scopes, validOIDCClientScopes, true)

	if len(duplicates) != 0 {
		errDeprecatedFunc()

		val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidEntryDuplicates, config.Clients[c].ID, attrOIDCScopes, strJoinAnd(duplicates)))
	}

	if utils.IsStringInSlice(oidc.GrantTypeClientCredentials, config.Clients[c].GrantTypes) {
		validateOIDCClientScopesClientCredentialsGrant(c, config, val)
	} else {
		if !utils.IsStringInSlice(oidc.ScopeOpenID, config.Clients[c].Scopes) {
			config.Clients[c].Scopes = append([]string{oidc.ScopeOpenID}, config.Clients[c].Scopes...)
		}

		if len(invalid) != 0 {
			val.Push(fmt.Errorf(errFmtOIDCClientInvalidEntries, config.Clients[c].ID, attrOIDCScopes, strJoinOr(validOIDCClientScopes), strJoinAnd(invalid)))
		}
	}

	if utils.IsStringSliceContainsAny([]string{oidc.ScopeOfflineAccess, oidc.ScopeOffline}, config.Clients[c].Scopes) &&
		!utils.IsStringSliceContainsAny(validOIDCClientResponseTypesRefreshToken, config.Clients[c].ResponseTypes) {
		errDeprecatedFunc()

		val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidRefreshTokenOptionWithoutCodeResponseType,
			config.Clients[c].ID, attrOIDCScopes,
			strJoinOr([]string{oidc.ScopeOfflineAccess, oidc.ScopeOffline}),
			strJoinOr(validOIDCClientResponseTypesRefreshToken)),
		)
	}
}

func validateOIDCClientScopesClientCredentialsGrant(c int, config *schema.OpenIDConnect, val *schema.StructValidator) {
	if len(config.Clients[c].GrantTypes) != 1 {
		return
	}

	invalid := validateListNotAllowed(config.Clients[c].Scopes, []string{oidc.ScopeOpenID, oidc.ScopeOffline, oidc.ScopeOfflineAccess})

	if len(invalid) > 0 {
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidEntriesClientCredentials, config.Clients[c].ID, strJoinAnd(config.Clients[c].Scopes), strJoinOr(invalid)))
	}
}

func validateOIDCClientResponseTypes(c int, config *schema.OpenIDConnect, val *schema.StructValidator, errDeprecatedFunc func()) {
	if len(config.Clients[c].ResponseTypes) == 0 {
		config.Clients[c].ResponseTypes = schema.DefaultOpenIDConnectClientConfiguration.ResponseTypes
	}

	invalid, duplicates := validateList(config.Clients[c].ResponseTypes, validOIDCClientResponseTypes, true)

	if len(invalid) != 0 {
		val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidEntries, config.Clients[c].ID, attrOIDCResponseTypes, strJoinOr(validOIDCClientResponseTypes), strJoinAnd(invalid)))
	}

	if len(duplicates) != 0 {
		errDeprecatedFunc()

		val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidEntryDuplicates, config.Clients[c].ID, attrOIDCResponseTypes, strJoinAnd(duplicates)))
	}
}

func validateOIDCClientResponseModes(c int, config *schema.OpenIDConnect, val *schema.StructValidator, errDeprecatedFunc func()) {
	if len(config.Clients[c].ResponseModes) == 0 {
		config.Clients[c].ResponseModes = schema.DefaultOpenIDConnectClientConfiguration.ResponseModes

		for _, responseType := range config.Clients[c].ResponseTypes {
			switch responseType {
			case oidc.ResponseTypeAuthorizationCodeFlow:
				if !utils.IsStringInSlice(oidc.ResponseModeQuery, config.Clients[c].ResponseModes) {
					config.Clients[c].ResponseModes = append(config.Clients[c].ResponseModes, oidc.ResponseModeQuery)
				}
			case oidc.ResponseTypeImplicitFlowIDToken, oidc.ResponseTypeImplicitFlowToken, oidc.ResponseTypeImplicitFlowBoth,
				oidc.ResponseTypeHybridFlowIDToken, oidc.ResponseTypeHybridFlowToken, oidc.ResponseTypeHybridFlowBoth:
				if !utils.IsStringInSlice(oidc.ResponseModeFragment, config.Clients[c].ResponseModes) {
					config.Clients[c].ResponseModes = append(config.Clients[c].ResponseModes, oidc.ResponseModeFragment)
				}
			}
		}
	}

	invalid, duplicates := validateList(config.Clients[c].ResponseModes, validOIDCClientResponseModes, true)

	if len(invalid) != 0 {
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidEntries, config.Clients[c].ID, attrOIDCResponseModes, strJoinOr(validOIDCClientResponseModes), strJoinAnd(invalid)))
	}

	if len(duplicates) != 0 {
		errDeprecatedFunc()

		val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidEntryDuplicates, config.Clients[c].ID, attrOIDCResponseModes, strJoinAnd(duplicates)))
	}
}

func validateOIDCClientGrantTypes(c int, config *schema.OpenIDConnect, val *schema.StructValidator, errDeprecatedFunc func()) {
	if len(config.Clients[c].GrantTypes) == 0 {
		validateOIDCClientGrantTypesSetDefaults(c, config)
	}

	validateOIDCClientGrantTypesCheckRelated(c, config, val, errDeprecatedFunc)

	invalid, duplicates := validateList(config.Clients[c].GrantTypes, validOIDCClientGrantTypes, true)

	if len(invalid) != 0 {
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidEntries, config.Clients[c].ID, attrOIDCGrantTypes, strJoinOr(validOIDCClientGrantTypes), strJoinAnd(invalid)))
	}

	if len(duplicates) != 0 {
		errDeprecatedFunc()

		val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidEntryDuplicates, config.Clients[c].ID, attrOIDCGrantTypes, strJoinAnd(duplicates)))
	}
}

func validateOIDCClientGrantTypesSetDefaults(c int, config *schema.OpenIDConnect) {
	for _, responseType := range config.Clients[c].ResponseTypes {
		switch responseType {
		case oidc.ResponseTypeAuthorizationCodeFlow:
			if !utils.IsStringInSlice(oidc.GrantTypeAuthorizationCode, config.Clients[c].GrantTypes) {
				config.Clients[c].GrantTypes = append(config.Clients[c].GrantTypes, oidc.GrantTypeAuthorizationCode)
			}
		case oidc.ResponseTypeImplicitFlowIDToken, oidc.ResponseTypeImplicitFlowToken, oidc.ResponseTypeImplicitFlowBoth:
			if !utils.IsStringInSlice(oidc.GrantTypeImplicit, config.Clients[c].GrantTypes) {
				config.Clients[c].GrantTypes = append(config.Clients[c].GrantTypes, oidc.GrantTypeImplicit)
			}
		case oidc.ResponseTypeHybridFlowIDToken, oidc.ResponseTypeHybridFlowToken, oidc.ResponseTypeHybridFlowBoth:
			if !utils.IsStringInSlice(oidc.GrantTypeAuthorizationCode, config.Clients[c].GrantTypes) {
				config.Clients[c].GrantTypes = append(config.Clients[c].GrantTypes, oidc.GrantTypeAuthorizationCode)
			}

			if !utils.IsStringInSlice(oidc.GrantTypeImplicit, config.Clients[c].GrantTypes) {
				config.Clients[c].GrantTypes = append(config.Clients[c].GrantTypes, oidc.GrantTypeImplicit)
			}
		}
	}
}

func validateOIDCClientGrantTypesCheckRelated(c int, config *schema.OpenIDConnect, val *schema.StructValidator, errDeprecatedFunc func()) {
	for _, grantType := range config.Clients[c].GrantTypes {
		switch grantType {
		case oidc.GrantTypeAuthorizationCode:
			if !utils.IsStringInSlice(oidc.ResponseTypeAuthorizationCodeFlow, config.Clients[c].ResponseTypes) && !utils.IsStringSliceContainsAny(validOIDCClientResponseTypesHybridFlow, config.Clients[c].ResponseTypes) {
				errDeprecatedFunc()

				val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidGrantTypeMatch, config.Clients[c].ID, grantType, "for either the authorization code or hybrid flow", strJoinOr(append([]string{oidc.ResponseTypeAuthorizationCodeFlow}, validOIDCClientResponseTypesHybridFlow...)), strJoinAnd(config.Clients[c].ResponseTypes)))
			}
		case oidc.GrantTypeImplicit:
			if !utils.IsStringSliceContainsAny(validOIDCClientResponseTypesImplicitFlow, config.Clients[c].ResponseTypes) && !utils.IsStringSliceContainsAny(validOIDCClientResponseTypesHybridFlow, config.Clients[c].ResponseTypes) {
				errDeprecatedFunc()

				val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidGrantTypeMatch, config.Clients[c].ID, grantType, "for either the implicit or hybrid flow", strJoinOr(append(append([]string{}, validOIDCClientResponseTypesImplicitFlow...), validOIDCClientResponseTypesHybridFlow...)), strJoinAnd(config.Clients[c].ResponseTypes)))
			}
		case oidc.GrantTypeClientCredentials:
			if config.Clients[c].Public {
				val.Push(fmt.Errorf(errFmtOIDCClientInvalidGrantTypePublic, config.Clients[c].ID, oidc.GrantTypeClientCredentials))
			}
		case oidc.GrantTypeRefreshToken:
			if !utils.IsStringSliceContainsAny([]string{oidc.ScopeOfflineAccess, oidc.ScopeOffline}, config.Clients[c].Scopes) {
				errDeprecatedFunc()

				val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidGrantTypeRefresh, config.Clients[c].ID))
			}

			if !utils.IsStringSliceContainsAny(validOIDCClientResponseTypesRefreshToken, config.Clients[c].ResponseTypes) {
				errDeprecatedFunc()

				val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidRefreshTokenOptionWithoutCodeResponseType,
					config.Clients[c].ID, attrOIDCGrantTypes,
					strJoinOr([]string{oidc.GrantTypeRefreshToken}),
					strJoinOr(validOIDCClientResponseTypesRefreshToken)),
				)
			}
		}
	}
}

func validateOIDCClientRedirectURIs(c int, config *schema.OpenIDConnect, val *schema.StructValidator, errDeprecatedFunc func()) {
	var (
		parsedRedirectURI *url.URL
		err               error
	)

	for _, redirectURI := range config.Clients[c].RedirectURIs {
		if redirectURI == oauth2InstalledApp {
			if config.Clients[c].Public {
				continue
			}

			val.Push(fmt.Errorf(errFmtOIDCClientRedirectURIPublic, config.Clients[c].ID, oauth2InstalledApp))

			continue
		}

		if parsedRedirectURI, err = url.Parse(redirectURI); err != nil {
			val.Push(fmt.Errorf(errFmtOIDCClientRedirectURICantBeParsed, config.Clients[c].ID, redirectURI, err))
			continue
		}

		if !parsedRedirectURI.IsAbs() || (!config.Clients[c].Public && parsedRedirectURI.Scheme == "") {
			val.Push(fmt.Errorf(errFmtOIDCClientRedirectURIAbsolute, config.Clients[c].ID, redirectURI))
		}
	}

	_, duplicates := validateList(config.Clients[c].RedirectURIs, nil, true)

	if len(duplicates) != 0 {
		errDeprecatedFunc()

		val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidEntryDuplicates, config.Clients[c].ID, attrOIDCRedirectURIs, strJoinAnd(duplicates)))
	}
}

func validateOIDCClientTokenEndpointAuth(c int, config *schema.OpenIDConnect, val *schema.StructValidator) {
	implcit := len(config.Clients[c].ResponseTypes) != 0 && utils.IsStringSliceContainsAll(config.Clients[c].ResponseTypes, validOIDCClientResponseTypesImplicitFlow)

	switch {
	case config.Clients[c].TokenEndpointAuthMethod == "":
		break
	case !utils.IsStringInSlice(config.Clients[c].TokenEndpointAuthMethod, validOIDCClientTokenEndpointAuthMethods):
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
			config.Clients[c].ID, attrOIDCTokenAuthMethod, strJoinOr(validOIDCClientTokenEndpointAuthMethods), config.Clients[c].TokenEndpointAuthMethod))
	case config.Clients[c].TokenEndpointAuthMethod == oidc.ClientAuthMethodNone && !config.Clients[c].Public && !implcit:
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidTokenEndpointAuthMethod,
			config.Clients[c].ID, strJoinOr(validOIDCClientTokenEndpointAuthMethodsConfidential), strJoinAnd(validOIDCClientResponseTypesImplicitFlow), config.Clients[c].TokenEndpointAuthMethod))
	case config.Clients[c].TokenEndpointAuthMethod != oidc.ClientAuthMethodNone && config.Clients[c].Public:
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidTokenEndpointAuthMethodPublic,
			config.Clients[c].ID, config.Clients[c].TokenEndpointAuthMethod))
	}

	switch config.Clients[c].TokenEndpointAuthMethod {
	case "":
		break
	case oidc.ClientAuthMethodClientSecretJWT:
		validateOIDCClientTokenEndpointAuthClientSecretJWT(c, config, val)
	case oidc.ClientAuthMethodPrivateKeyJWT:
		validateOIDCClientTokenEndpointAuthPublicKeyJWT(config.Clients[c], val)
	}
}

func validateOIDCClientTokenEndpointAuthClientSecretJWT(c int, config *schema.OpenIDConnect, val *schema.StructValidator) {
	switch {
	case config.Clients[c].TokenEndpointAuthSigningAlg == "":
		config.Clients[c].TokenEndpointAuthSigningAlg = oidc.SigningAlgHMACUsingSHA256
	case !utils.IsStringInSlice(config.Clients[c].TokenEndpointAuthSigningAlg, validOIDCClientTokenEndpointAuthSigAlgsClientSecretJWT):
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidTokenEndpointAuthSigAlg, config.Clients[c].ID, strJoinOr(validOIDCClientTokenEndpointAuthSigAlgsClientSecretJWT), config.Clients[c].TokenEndpointAuthMethod))
	}
}

func validateOIDCClientTokenEndpointAuthPublicKeyJWT(config schema.OpenIDConnectClient, val *schema.StructValidator) {
	switch {
	case config.TokenEndpointAuthSigningAlg == "":
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidTokenEndpointAuthSigAlgMissingPrivateKeyJWT, config.ID))
	case !utils.IsStringInSlice(config.TokenEndpointAuthSigningAlg, validOIDCIssuerJWKSigningAlgs):
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidTokenEndpointAuthSigAlg, config.ID, strJoinOr(validOIDCIssuerJWKSigningAlgs), config.TokenEndpointAuthMethod))
	}

	if config.PublicKeys.URI == nil {
		if len(config.PublicKeys.Values) == 0 {
			val.Push(fmt.Errorf(errFmtOIDCClientInvalidPublicKeysPrivateKeyJWT, config.ID))
		} else if len(config.Discovery.RequestObjectSigningAlgs) != 0 && !utils.IsStringInSlice(config.TokenEndpointAuthSigningAlg, config.Discovery.RequestObjectSigningAlgs) {
			val.Push(fmt.Errorf(errFmtOIDCClientInvalidTokenEndpointAuthSigAlgReg, config.ID, strJoinOr(config.Discovery.RequestObjectSigningAlgs), config.TokenEndpointAuthMethod))
		}
	}
}

func validateOIDDClientSigningAlgs(c int, config *schema.OpenIDConnect, val *schema.StructValidator) {
	switch config.Clients[c].UserinfoSigningKeyID {
	case "":
		if config.Clients[c].UserinfoSigningAlg == "" {
			config.Clients[c].UserinfoSigningAlg = schema.DefaultOpenIDConnectClientConfiguration.UserinfoSigningAlg
		} else if config.Clients[c].UserinfoSigningAlg != oidc.SigningAlgNone && !utils.IsStringInSlice(config.Clients[c].UserinfoSigningAlg, config.Discovery.ResponseObjectSigningAlgs) {
			val.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
				config.Clients[c].ID, attrOIDCUsrSigAlg, strJoinOr(append(config.Discovery.ResponseObjectSigningAlgs, oidc.SigningAlgNone)), config.Clients[c].UserinfoSigningAlg))
		}
	default:
		if !utils.IsStringInSlice(config.Clients[c].UserinfoSigningKeyID, config.Discovery.ResponseObjectSigningKeyIDs) {
			val.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
				config.Clients[c].ID, attrOIDCUsrSigKID, strJoinOr(config.Discovery.ResponseObjectSigningKeyIDs), config.Clients[c].UserinfoSigningKeyID))
		} else {
			config.Clients[c].UserinfoSigningAlg = getResponseObjectAlgFromKID(config, config.Clients[c].UserinfoSigningKeyID, config.Clients[c].UserinfoSigningAlg)
		}
	}

	switch config.Clients[c].IDTokenSigningKeyID {
	case "":
		if config.Clients[c].IDTokenSigningAlg == "" {
			config.Clients[c].IDTokenSigningAlg = schema.DefaultOpenIDConnectClientConfiguration.IDTokenSigningAlg
		} else if !utils.IsStringInSlice(config.Clients[c].IDTokenSigningAlg, config.Discovery.ResponseObjectSigningAlgs) {
			val.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
				config.Clients[c].ID, attrOIDCIDTokenSigAlg, strJoinOr(config.Discovery.ResponseObjectSigningAlgs), config.Clients[c].IDTokenSigningAlg))
		}
	default:
		if !utils.IsStringInSlice(config.Clients[c].IDTokenSigningKeyID, config.Discovery.ResponseObjectSigningKeyIDs) {
			val.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
				config.Clients[c].ID, attrOIDCIDTokenSigKID, strJoinOr(config.Discovery.ResponseObjectSigningKeyIDs), config.Clients[c].IDTokenSigningKeyID))
		} else {
			config.Clients[c].IDTokenSigningAlg = getResponseObjectAlgFromKID(config, config.Clients[c].IDTokenSigningKeyID, config.Clients[c].IDTokenSigningAlg)
		}
	}
}
