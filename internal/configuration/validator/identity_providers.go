package validator

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"net/url"
	"sort"
	"strconv"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateIdentityProviders validates and updates the IdentityProviders configuration.
func ValidateIdentityProviders(ctx *ValidateCtx, config *schema.IdentityProviders, validator *schema.StructValidator) {
	validateOIDC(ctx, config.OIDC, validator)
}

func validateOIDC(ctx *ValidateCtx, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	if config == nil {
		return
	}

	setOIDCDefaults(config)

	validateOIDCIssuer(config, validator)
	validateOIDCAuthorizationPolicies(config, validator)
	validateOIDCLifespans(config, validator)

	sort.Sort(oidc.SortedSigningAlgs(config.Discovery.ResponseObjectSigningAlgs))

	switch {
	case config.MinimumParameterEntropy == -1:
		validator.PushWarning(fmt.Errorf(errFmtOIDCProviderInsecureDisabledParameterEntropy))
	case config.MinimumParameterEntropy <= 0:
		config.MinimumParameterEntropy = oauthelia2.MinParameterEntropy
	case config.MinimumParameterEntropy < oauthelia2.MinParameterEntropy:
		validator.PushWarning(fmt.Errorf(errFmtOIDCProviderInsecureParameterEntropyUnsafe, oauthelia2.MinParameterEntropy, config.MinimumParameterEntropy))
	}

	switch config.EnforcePKCE {
	case "always", "never", "public_clients_only":
		break
	default:
		validator.Push(fmt.Errorf(errFmtOIDCProviderEnforcePKCEInvalidValue, config.EnforcePKCE))
	}

	validateOIDCOptionsCORS(config, validator)

	if len(config.Clients) == 0 {
		validator.Push(fmt.Errorf(errFmtOIDCProviderNoClientsConfigured))
	} else {
		validateOIDCClients(ctx, config, validator)
	}
}

func validateOIDCAuthorizationPolicies(config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	config.Discovery.AuthorizationPolicies = []string{policyOneFactor, policyTwoFactor}

	for name, policy := range config.AuthorizationPolicies {
		add := true

		switch name {
		case "":
			validator.Push(fmt.Errorf(errFmtOIDCPolicyInvalidName))

			add = false
		case policyOneFactor, policyTwoFactor, policyDeny:
			validator.Push(fmt.Errorf(errFmtOIDCPolicyInvalidNameStandard, name, "name", utils.StringJoinAnd([]string{policyOneFactor, policyTwoFactor, policyDeny}), name))

			add = false
		}

		switch policy.DefaultPolicy {
		case "":
			policy.DefaultPolicy = schema.DefaultOpenIDConnectPolicyConfiguration.DefaultPolicy
		case policyOneFactor, policyTwoFactor, policyDeny:
			break
		default:
			validator.Push(fmt.Errorf(errFmtOIDCPolicyInvalidDefaultPolicy, name, utils.StringJoinAnd([]string{policyOneFactor, policyTwoFactor, policyDeny}), policy.DefaultPolicy))
		}

		if len(policy.Rules) == 0 {
			validator.Push(fmt.Errorf(errFmtOIDCPolicyMissingOption, name, "rules"))
		}

		for i, rule := range policy.Rules {
			switch rule.Policy {
			case "":
				policy.Rules[i].Policy = schema.DefaultOpenIDConnectPolicyConfiguration.DefaultPolicy
			case policyOneFactor, policyTwoFactor, policyDeny:
				break
			default:
				validator.Push(fmt.Errorf(errFmtOIDCPolicyRuleInvalidPolicy, name, i+1, utils.StringJoinAnd([]string{policyOneFactor, policyTwoFactor, policyDeny}), rule.Policy))
			}

			if len(rule.Subjects) == 0 {
				validator.Push(fmt.Errorf(errFmtOIDCPolicyRuleMissingOption, name, i+1, "subject"))
			}
		}

		config.AuthorizationPolicies[name] = policy

		if add {
			config.Discovery.AuthorizationPolicies = append(config.Discovery.AuthorizationPolicies, name)
		}
	}
}

func validateOIDCLifespans(config *schema.IdentityProvidersOpenIDConnect, _ *schema.StructValidator) {
	for name := range config.Lifespans.Custom {
		config.Discovery.Lifespans = append(config.Discovery.Lifespans, name)
	}
}

func validateOIDCIssuer(config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	switch {
	case len(config.JSONWebKeys) != 0 && (config.IssuerPrivateKey != nil || config.IssuerCertificateChain.HasCertificates()):
		validator.Push(fmt.Errorf("identity_providers: oidc: option `jwks` must not be configured at the same time as 'issuer_private_key' or 'issuer_certificate_chain'"))
	case config.IssuerPrivateKey != nil:
		validateOIDCIssuerPrivateKey(config)

		fallthrough
	case len(config.JSONWebKeys) != 0:
		validateOIDCIssuerJSONWebKeys(config, validator)
		validateOIDDIssuerSigningAlgsDiscovery(config, validator)
	default:
		validator.Push(fmt.Errorf(errFmtOIDCProviderNoPrivateKey))
	}
}

func validateOIDDIssuerSigningAlgsDiscovery(config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	config.DiscoverySignedResponseAlg, config.DiscoverySignedResponseKeyID = validateOIDCAlgKIDDefault(config, config.DiscoverySignedResponseAlg, config.DiscoverySignedResponseKeyID, schema.DefaultOpenIDConnectConfiguration.DiscoverySignedResponseAlg)

	switch config.DiscoverySignedResponseKeyID {
	case "":
		switch config.DiscoverySignedResponseAlg {
		case "", oidc.SigningAlgNone, oidc.SigningAlgRSAUsingSHA256:
			break
		default:
			if !utils.IsStringInSlice(config.DiscoverySignedResponseAlg, config.Discovery.ResponseObjectSigningAlgs) {
				validator.Push(fmt.Errorf(errFmtOIDCProviderInvalidValue, attrOIDCDiscoSigAlg, utils.StringJoinOr(append(config.Discovery.ResponseObjectSigningAlgs, oidc.SigningAlgNone)), config.DiscoverySignedResponseAlg))
			}
		}
	default:
		if !utils.IsStringInSlice(config.DiscoverySignedResponseKeyID, config.Discovery.ResponseObjectSigningKeyIDs) {
			validator.Push(fmt.Errorf(errFmtOIDCProviderInvalidValue, attrOIDCDiscoSigKID, utils.StringJoinOr(config.Discovery.ResponseObjectSigningKeyIDs), config.DiscoverySignedResponseKeyID))
		} else {
			config.DiscoverySignedResponseAlg = getResponseObjectAlgFromKID(config, config.DiscoverySignedResponseKeyID, config.DiscoverySignedResponseAlg)
		}
	}
}

func validateOIDCIssuerPrivateKey(config *schema.IdentityProvidersOpenIDConnect) {
	config.JSONWebKeys = append([]schema.JWK{{
		Algorithm:        oidc.SigningAlgRSAUsingSHA256,
		Use:              oidc.KeyUseSignature,
		Key:              config.IssuerPrivateKey,
		CertificateChain: config.IssuerCertificateChain,
	}}, config.JSONWebKeys...)
}

func validateOIDCIssuerJSONWebKeys(config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	var (
		props *JWKProperties
		err   error
	)

	config.Discovery.ResponseObjectSigningKeyIDs = make([]string, len(config.JSONWebKeys))
	config.Discovery.DefaultKeyIDs = map[string]string{}

	for i := 0; i < len(config.JSONWebKeys); i++ {
		if key, ok := config.JSONWebKeys[i].Key.(*rsa.PrivateKey); ok && key.PublicKey.N == nil {
			validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysInvalid, i+1))

			continue
		}

		if props, err = schemaJWKGetProperties(config.JSONWebKeys[i]); err != nil {
			validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysProperties, i+1, config.JSONWebKeys[i].KeyID, err))

			continue
		}

		switch n := len(config.JSONWebKeys[i].KeyID); {
		case n == 0:
			if config.JSONWebKeys[i].KeyID, err = jwkCalculateKID(config.JSONWebKeys[i].Key, props, config.JSONWebKeys[i].Algorithm); err != nil {
				validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysCalcThumbprint, i+1, err))

				continue
			}
		case n > 100:
			validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysKeyIDLength, i+1, config.JSONWebKeys[i].KeyID))
		}

		if config.JSONWebKeys[i].KeyID != "" && utils.IsStringInSlice(config.JSONWebKeys[i].KeyID, config.Discovery.ResponseObjectSigningKeyIDs) {
			validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysAttributeNotUnique, i+1, config.JSONWebKeys[i].KeyID, attrOIDCKeyID))
		}

		config.Discovery.ResponseObjectSigningKeyIDs[i] = config.JSONWebKeys[i].KeyID

		if !reOpenIDConnectKID.MatchString(config.JSONWebKeys[i].KeyID) {
			validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysKeyIDNotValid, i+1, config.JSONWebKeys[i].KeyID))
		}

		validateOIDCIssuerPrivateKeysUseAlg(i, props, config, validator)
		validateOIDCIssuerPrivateKeyPair(i, config, validator)
	}

	if len(config.Discovery.ResponseObjectSigningAlgs) != 0 && !utils.IsStringInSlice(oidc.SigningAlgRSAUsingSHA256, config.Discovery.ResponseObjectSigningAlgs) {
		validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysNoRS256, oidc.SigningAlgRSAUsingSHA256, utils.StringJoinAnd(config.Discovery.ResponseObjectSigningAlgs)))
	}
}

func validateOIDCIssuerPrivateKeysUseAlg(i int, props *JWKProperties, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	switch config.JSONWebKeys[i].Use {
	case "":
		config.JSONWebKeys[i].Use = props.Use
	case oidc.KeyUseSignature:
		break
	default:
		validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysInvalidOptionOneOf, i+1, config.JSONWebKeys[i].KeyID, attrOIDCKeyUse, utils.StringJoinOr([]string{oidc.KeyUseSignature}), config.JSONWebKeys[i].Use))
	}

	switch {
	case config.JSONWebKeys[i].Algorithm == "":
		config.JSONWebKeys[i].Algorithm = props.Algorithm

		fallthrough
	case utils.IsStringInSlice(config.JSONWebKeys[i].Algorithm, validOIDCIssuerJWKSigningAlgs):
		if config.JSONWebKeys[i].KeyID != "" && config.JSONWebKeys[i].Algorithm != "" {
			if _, ok := config.Discovery.DefaultKeyIDs[config.JSONWebKeys[i].Algorithm]; !ok {
				config.Discovery.DefaultKeyIDs[config.JSONWebKeys[i].Algorithm] = config.JSONWebKeys[i].KeyID
			}
		}
	default:
		validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysInvalidOptionOneOf, i+1, config.JSONWebKeys[i].KeyID, attrOIDCAlgorithm, utils.StringJoinOr(validOIDCIssuerJWKSigningAlgs), config.JSONWebKeys[i].Algorithm))
	}

	if config.JSONWebKeys[i].Algorithm != "" {
		if !utils.IsStringInSlice(config.JSONWebKeys[i].Algorithm, config.Discovery.ResponseObjectSigningAlgs) {
			config.Discovery.ResponseObjectSigningAlgs = append(config.Discovery.ResponseObjectSigningAlgs, config.JSONWebKeys[i].Algorithm)
		}
	}
}

func validateOIDCIssuerPrivateKeyPair(i int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	var (
		checkEqualKey bool
		err           error
	)

	switch key := config.JSONWebKeys[i].Key.(type) {
	case *rsa.PrivateKey:
		checkEqualKey = true

		if key.Size() < 256 {
			checkEqualKey = false

			validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysRSAKeyLessThan2048Bits, i+1, config.JSONWebKeys[i].KeyID, key.Size()*8))
		}
	case *ecdsa.PrivateKey:
		checkEqualKey = true
	default:
		validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysKeyNotRSAOrECDSA, i+1, config.JSONWebKeys[i].KeyID, key))
	}

	if config.JSONWebKeys[i].CertificateChain.HasCertificates() {
		if checkEqualKey && !config.JSONWebKeys[i].CertificateChain.EqualKey(config.JSONWebKeys[i].Key) {
			validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysKeyCertificateMismatch, i+1, config.JSONWebKeys[i].KeyID))
		}

		if err = config.JSONWebKeys[i].CertificateChain.Validate(); err != nil {
			validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysCertificateChainInvalid, i+1, config.JSONWebKeys[i].KeyID, err))
		}
	}
}

func setOIDCDefaults(config *schema.IdentityProvidersOpenIDConnect) {
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

func validateOIDCOptionsCORS(config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	validateOIDCOptionsCORSAllowedOrigins(config, validator)

	if config.CORS.AllowedOriginsFromClientRedirectURIs {
		validateOIDCOptionsCORSAllowedOriginsFromClientRedirectURIs(config)
	}

	validateOIDCOptionsCORSEndpoints(config, validator)
}

func validateOIDCOptionsCORSAllowedOrigins(config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	for _, origin := range config.CORS.AllowedOrigins {
		if origin.String() == "*" {
			if len(config.CORS.AllowedOrigins) != 1 {
				validator.Push(fmt.Errorf(errFmtOIDCCORSInvalidOriginWildcard))
			}

			if config.CORS.AllowedOriginsFromClientRedirectURIs {
				validator.Push(fmt.Errorf(errFmtOIDCCORSInvalidOriginWildcardWithClients))
			}

			continue
		}

		if origin.Path != "" {
			validator.Push(fmt.Errorf(errFmtOIDCCORSInvalidOrigin, origin.String(), "path"))
		}

		if origin.RawQuery != "" {
			validator.Push(fmt.Errorf(errFmtOIDCCORSInvalidOrigin, origin.String(), "query string"))
		}
	}
}

func validateOIDCOptionsCORSAllowedOriginsFromClientRedirectURIs(config *schema.IdentityProvidersOpenIDConnect) {
	for _, client := range config.Clients {
		for _, redirectURI := range client.RedirectURIs {
			uri, err := url.ParseRequestURI(redirectURI)
			if err != nil || (uri.Scheme != schemeHTTP && uri.Scheme != schemeHTTPS) || uri.Hostname() == "localhost" {
				continue
			}

			origin := utils.OriginFromURL(uri)

			if !utils.IsURLInSlice(origin, config.CORS.AllowedOrigins) {
				config.CORS.AllowedOrigins = append(config.CORS.AllowedOrigins, origin)
			}
		}
	}
}

func validateOIDCOptionsCORSEndpoints(config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	for _, endpoint := range config.CORS.Endpoints {
		if !utils.IsStringInSlice(endpoint, validOIDCCORSEndpoints) {
			validator.Push(fmt.Errorf(errFmtOIDCCORSInvalidEndpoint, endpoint, utils.StringJoinOr(validOIDCCORSEndpoints)))
		}
	}
}

func validateOIDCClients(ctx *ValidateCtx, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	var (
		errDeprecated bool

		clientIDs, duplicateClientIDs, blankClientIDs []string
	)

	ctx.cacheSectorIdentifierURIs = map[string][]string{}

	errDeprecatedFunc := func() { errDeprecated = true }

	for c, client := range config.Clients {
		n := len(client.ID)

		switch {
		case n == 0:
			blankClientIDs = append(blankClientIDs, "#"+strconv.Itoa(c+1))
		case n > 100:
			validator.Push(fmt.Errorf(errFmtOIDCClientIDTooLong, client.ID, n))
		case !reRFC3986Unreserved.MatchString(client.ID):
			validator.Push(fmt.Errorf(errFmtOIDCClientIDInvalidCharacters, client.ID))
		default:
			if client.Name == "" {
				config.Clients[c].Name = client.ID
			}

			if utils.IsStringInSlice(client.ID, clientIDs) {
				if !utils.IsStringInSlice(client.ID, duplicateClientIDs) {
					duplicateClientIDs = append(duplicateClientIDs, client.ID)
				}
			} else {
				clientIDs = append(clientIDs, client.ID)
			}
		}

		validateOIDCClient(ctx, c, config, validator, errDeprecatedFunc)
	}

	if errDeprecated {
		validator.PushWarning(fmt.Errorf(errFmtOIDCClientsDeprecated))
	}

	if len(blankClientIDs) != 0 {
		validator.Push(fmt.Errorf(errFmtOIDCClientsWithEmptyID, utils.StringJoinBuild(", ", "or", "", blankClientIDs)))
	}

	if len(duplicateClientIDs) != 0 {
		validator.Push(fmt.Errorf(errFmtOIDCClientsDuplicateID, utils.StringJoinOr(duplicateClientIDs)))
	}

	ctx.cacheSectorIdentifierURIs = nil
}

func validateOIDCClient(ctx *ValidateCtx, c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator, errDeprecatedFunc func()) {
	ccg := utils.IsStringInSlice(oidc.GrantTypeClientCredentials, config.Clients[c].GrantTypes)

	switch {
	case ccg:
		if config.Clients[c].AuthorizationPolicy == "" {
			config.Clients[c].AuthorizationPolicy = policyOneFactor
		} else if config.Clients[c].AuthorizationPolicy != policyOneFactor {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue, config.Clients[c].ID, "authorization_policy", utils.StringJoinOr([]string{policyOneFactor}), config.Clients[c].AuthorizationPolicy))
		}
	case config.Clients[c].AuthorizationPolicy == "":
		config.Clients[c].AuthorizationPolicy = schema.DefaultOpenIDConnectClientConfiguration.AuthorizationPolicy
	case utils.IsStringInSlice(config.Clients[c].AuthorizationPolicy, config.Discovery.AuthorizationPolicies):
		break
	default:
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue, config.Clients[c].ID, "authorization_policy", utils.StringJoinOr(config.Discovery.AuthorizationPolicies), config.Clients[c].AuthorizationPolicy))
	}

	switch {
	case config.Clients[c].Lifespan == "", utils.IsStringInSlice(config.Clients[c].Lifespan, config.Discovery.Lifespans):
		break
	default:
		if len(config.Discovery.Lifespans) == 0 {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidLifespan, config.Clients[c].ID, config.Clients[c].Lifespan))
		} else {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue, config.Clients[c].ID, "lifespan", utils.StringJoinOr(config.Discovery.Lifespans), config.Clients[c].Lifespan))
		}
	}

	switch config.Clients[c].PKCEChallengeMethod {
	case "", oidc.PKCEChallengeMethodPlain, oidc.PKCEChallengeMethodSHA256:
		break
	default:
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue, config.Clients[c].ID, attrOIDCPKCEChallengeMethod, utils.StringJoinOr([]string{oidc.PKCEChallengeMethodPlain, oidc.PKCEChallengeMethodSHA256}), config.Clients[c].PKCEChallengeMethod))
	}

	switch config.Clients[c].RequestedAudienceMode {
	case "":
		config.Clients[c].RequestedAudienceMode = schema.DefaultOpenIDConnectClientConfiguration.RequestedAudienceMode
	case oidc.ClientRequestedAudienceModeExplicit.String(), oidc.ClientRequestedAudienceModeImplicit.String():
		break
	default:
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue, config.Clients[c].ID, attrOIDCRequestedAudienceMode, utils.StringJoinOr([]string{oidc.ClientRequestedAudienceModeExplicit.String(), oidc.ClientRequestedAudienceModeImplicit.String()}), config.Clients[c].RequestedAudienceMode))
	}

	setDefaults := validateOIDCClientScopesSpecialBearerAuthz(c, config, ccg, validator)

	validateOIDCClientConsentMode(c, config, validator, setDefaults)

	validateOIDCClientScopes(c, config, validator, ccg, errDeprecatedFunc)
	validateOIDCClientResponseTypes(c, config, validator, setDefaults, errDeprecatedFunc)
	validateOIDCClientResponseModes(c, config, validator, setDefaults, errDeprecatedFunc)
	validateOIDCClientGrantTypes(c, config, validator, setDefaults, errDeprecatedFunc)
	validateOIDCClientRedirectURIs(c, config, validator, errDeprecatedFunc)
	validateOIDCClientRequestURIs(c, config, validator)

	validateOIDDClientSigningAlgs(c, config, validator)

	validateOIDCClientSectorIdentifier(ctx, c, config, validator, errDeprecatedFunc)

	validateOIDCClientPublicKeys(c, config, validator)
	validateOIDCClientTokenEndpointAuth(c, config, validator)
}

func validateOIDCClientPublicKeys(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	switch {
	case config.Clients[c].JSONWebKeysURI != nil && len(config.Clients[c].JSONWebKeys) != 0:
		validator.Push(fmt.Errorf(errFmtOIDCClientPublicKeysBothURIAndValuesConfigured, config.Clients[c].ID))
	case config.Clients[c].JSONWebKeysURI != nil:
		if config.Clients[c].JSONWebKeysURI.Scheme != schemeHTTPS {
			validator.Push(fmt.Errorf(errFmtOIDCClientPublicKeysURIInvalidScheme, config.Clients[c].ID, config.Clients[c].JSONWebKeysURI.Scheme))
		}
	case len(config.Clients[c].JSONWebKeys) != 0:
		validateOIDCClientJSONWebKeysList(c, config, validator)
	}
}

//nolint:gocyclo
func validateOIDCClientJSONWebKeysList(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	var (
		props *JWKProperties
		err   error
	)

	for i := 0; i < len(config.Clients[c].JSONWebKeys); i++ {
		if config.Clients[c].JSONWebKeys[i].KeyID == "" {
			validator.Push(fmt.Errorf(errFmtOIDCClientPublicKeysInvalidOptionMissingOneOf, config.Clients[c].ID, i+1, attrOIDCKeyID))
		}

		if props, err = schemaJWKGetProperties(config.Clients[c].JSONWebKeys[i]); err != nil {
			validator.Push(fmt.Errorf(errFmtOIDCClientPublicKeysProperties, config.Clients[c].ID, i+1, config.Clients[c].JSONWebKeys[i].KeyID, err))

			continue
		}

		validateOIDCClientJSONWebKeysListKeyUseAlg(c, i, props, config, validator)

		var checkEqualKey bool

		switch key := config.Clients[c].JSONWebKeys[i].Key.(type) {
		case nil:
			validator.Push(fmt.Errorf(errFmtOIDCClientPublicKeysInvalidOptionMissingOneOf, config.Clients[c].ID, i+1, attrOIDCKey))
		case *rsa.PublicKey:
			checkEqualKey = true

			if key.N == nil {
				checkEqualKey = false

				validator.Push(fmt.Errorf(errFmtOIDCClientPublicKeysKeyMalformed, config.Clients[c].ID, i+1))
			} else if key.Size() < 256 {
				checkEqualKey = false

				validator.Push(fmt.Errorf(errFmtOIDCClientPublicKeysRSAKeyLessThan2048Bits, config.Clients[c].ID, i+1, config.Clients[c].JSONWebKeys[i].KeyID, key.Size()*8))
			}
		case *ecdsa.PublicKey:
			checkEqualKey = true
		default:
			validator.Push(fmt.Errorf(errFmtOIDCClientPublicKeysKeyNotRSAOrECDSA, config.Clients[c].ID, i+1, config.Clients[c].JSONWebKeys[i].KeyID, key))
		}

		if config.Clients[c].JSONWebKeys[i].CertificateChain.HasCertificates() {
			if checkEqualKey && !config.Clients[c].JSONWebKeys[i].CertificateChain.EqualKey(config.Clients[c].JSONWebKeys[i].Key) {
				validator.Push(fmt.Errorf(errFmtOIDCClientPublicKeysCertificateChainKeyMismatch, config.Clients[c].ID, i+1, config.Clients[c].JSONWebKeys[i].KeyID))
			}

			if err = config.Clients[c].JSONWebKeys[i].CertificateChain.Validate(); err != nil {
				validator.Push(fmt.Errorf(errFmtOIDCClientPublicKeysCertificateChainInvalid, config.Clients[c].ID, i+1, config.Clients[c].JSONWebKeys[i].KeyID, err))
			}
		}
	}

	if config.Clients[c].RequestObjectSigningAlg != "" && config.Clients[c].JSONWebKeysURI == nil && !utils.IsStringInSlice(config.Clients[c].RequestObjectSigningAlg, config.Clients[c].Discovery.RequestObjectSigningAlgs) {
		validator.Push(fmt.Errorf(errFmtOIDCClientPublicKeysROSAMissingAlgorithm, config.Clients[c].ID, utils.StringJoinOr(config.Clients[c].Discovery.RequestObjectSigningAlgs)))
	}
}

func validateOIDCClientJSONWebKeysListKeyUseAlg(c, i int, props *JWKProperties, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	switch config.Clients[c].JSONWebKeys[i].Use {
	case "":
		config.Clients[c].JSONWebKeys[i].Use = props.Use
	case oidc.KeyUseSignature:
		break
	default:
		validator.Push(fmt.Errorf(errFmtOIDCClientPublicKeysInvalidOptionOneOf, config.Clients[c].ID, i+1, config.Clients[c].JSONWebKeys[i].KeyID, attrOIDCKeyUse, utils.StringJoinOr([]string{oidc.KeyUseSignature}), config.Clients[c].JSONWebKeys[i].Use))
	}

	switch {
	case config.Clients[c].JSONWebKeys[i].Algorithm == "":
		config.Clients[c].JSONWebKeys[i].Algorithm = props.Algorithm
	case utils.IsStringInSlice(config.Clients[c].JSONWebKeys[i].Algorithm, validOIDCIssuerJWKSigningAlgs):
		break
	default:
		validator.Push(fmt.Errorf(errFmtOIDCClientPublicKeysInvalidOptionOneOf, config.Clients[c].ID, i+1, config.Clients[c].JSONWebKeys[i].KeyID, attrOIDCAlgorithm, utils.StringJoinOr(validOIDCIssuerJWKSigningAlgs), config.Clients[c].JSONWebKeys[i].Algorithm))
	}

	if config.Clients[c].JSONWebKeys[i].Algorithm != "" {
		if !utils.IsStringInSlice(config.Clients[c].JSONWebKeys[i].Algorithm, config.Discovery.RequestObjectSigningAlgs) {
			config.Discovery.RequestObjectSigningAlgs = append(config.Discovery.RequestObjectSigningAlgs, config.Clients[c].JSONWebKeys[i].Algorithm)
		}

		if !utils.IsStringInSlice(config.Clients[c].JSONWebKeys[i].Algorithm, config.Clients[c].Discovery.RequestObjectSigningAlgs) {
			config.Clients[c].Discovery.RequestObjectSigningAlgs = append(config.Clients[c].Discovery.RequestObjectSigningAlgs, config.Clients[c].JSONWebKeys[i].Algorithm)
		}
	}
}

func validateOIDCClientSectorIdentifier(ctx *ValidateCtx, c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator, errDeprecatedFunc func()) {
	if config.Clients[c].SectorIdentifierURI == nil {
		return
	}

	if config.Clients[c].SectorIdentifierURI.String() == "" {
		config.Clients[c].SectorIdentifierURI = nil

		return
	}

	valid := true

	if !config.Clients[c].SectorIdentifierURI.IsAbs() {
		errDeprecatedFunc()
		validator.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifierAbsolute, config.Clients[c].ID, config.Clients[c].SectorIdentifierURI.String()))

		valid = false
	} else if config.Clients[c].SectorIdentifierURI.Scheme != schemeHTTPS {
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifierScheme, config.Clients[c].ID, config.Clients[c].SectorIdentifierURI.String(), config.Clients[c].SectorIdentifierURI.Scheme))

		valid = false
	}

	if config.Clients[c].SectorIdentifierURI.Fragment != "" {
		errDeprecatedFunc()
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifier, config.Clients[c].ID, config.Clients[c].SectorIdentifierURI.String(), "fragment", "fragment", config.Clients[c].SectorIdentifierURI.Fragment))

		valid = false
	}

	if config.Clients[c].SectorIdentifierURI.User != nil {
		if config.Clients[c].SectorIdentifierURI.User.Username() != "" {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifier, config.Clients[c].ID, config.Clients[c].SectorIdentifierURI.String(), "username", "username", config.Clients[c].SectorIdentifierURI.User.Username()))
		}

		if password, set := config.Clients[c].SectorIdentifierURI.User.Password(); set {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifier, config.Clients[c].ID, config.Clients[c].SectorIdentifierURI.String(), "password", "password", password))
		}

		valid = false
	}

	if valid {
		if err := oidc.ValidateSectorIdentifierURI(ctx, ctx.cacheSectorIdentifierURIs, config.Clients[c].SectorIdentifierURI, config.Clients[c].RedirectURIs); err != nil {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifierRedirect, config.Clients[c].ID, config.Clients[c].SectorIdentifierURI.String(), err))
		}
	}
}

func validateOIDCClientConsentMode(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator, setDefaults bool) {
	switch {
	case utils.IsStringInSlice(config.Clients[c].ConsentMode, []string{"", auto}):
		if !setDefaults {
			break
		}

		if config.Clients[c].ConsentPreConfiguredDuration != nil {
			config.Clients[c].ConsentMode = oidc.ClientConsentModePreConfigured.String()
		} else {
			config.Clients[c].ConsentMode = oidc.ClientConsentModeExplicit.String()
		}
	case utils.IsStringInSlice(config.Clients[c].ConsentMode, validOIDCClientConsentModes):
		break
	default:
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidConsentMode, config.Clients[c].ID, utils.StringJoinOr(append(validOIDCClientConsentModes, auto)), config.Clients[c].ConsentMode))
	}

	if config.Clients[c].ConsentMode == oidc.ClientConsentModePreConfigured.String() && config.Clients[c].ConsentPreConfiguredDuration == nil {
		config.Clients[c].ConsentPreConfiguredDuration = schema.DefaultOpenIDConnectClientConfiguration.ConsentPreConfiguredDuration
	}
}

func validateOIDCClientScopes(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator, ccg bool, errDeprecatedFunc func()) {
	if len(config.Clients[c].Scopes) == 0 && !ccg {
		config.Clients[c].Scopes = schema.DefaultOpenIDConnectClientConfiguration.Scopes
	}

	invalid, duplicates := validateList(config.Clients[c].Scopes, validOIDCClientScopes, true)

	if len(duplicates) != 0 {
		errDeprecatedFunc()

		validator.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidEntryDuplicates, config.Clients[c].ID, attrOIDCScopes, utils.StringJoinAnd(duplicates)))
	}

	if ccg {
		validateOIDCClientScopesClientCredentialsGrant(c, config, validator)
	} else if len(invalid) != 0 {
		validator.PushWarning(fmt.Errorf(errFmtOIDCClientUnknownScopeEntries, config.Clients[c].ID, attrOIDCScopes, utils.StringJoinOr(validOIDCClientScopes), utils.StringJoinAnd(invalid)))
	}

	if utils.IsStringSliceContainsAny([]string{oidc.ScopeOfflineAccess, oidc.ScopeOffline}, config.Clients[c].Scopes) &&
		!utils.IsStringSliceContainsAny(validOIDCClientResponseTypesRefreshToken, config.Clients[c].ResponseTypes) {
		errDeprecatedFunc()

		validator.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidRefreshTokenOptionWithoutCodeResponseType,
			config.Clients[c].ID, attrOIDCScopes,
			utils.StringJoinOr([]string{oidc.ScopeOfflineAccess, oidc.ScopeOffline}),
			utils.StringJoinOr(validOIDCClientResponseTypesRefreshToken)),
		)
	}
}

//nolint:gocyclo
func validateOIDCClientScopesSpecialBearerAuthz(c int, config *schema.IdentityProvidersOpenIDConnect, ccg bool, validator *schema.StructValidator) bool {
	if !utils.IsStringInSlice(oidc.ScopeAutheliaBearerAuthz, config.Clients[c].Scopes) {
		return true
	}

	if !config.Discovery.BearerAuthorization {
		config.Discovery.BearerAuthorization = true
	}

	if !utils.IsStringSliceContainsAll(config.Clients[c].Scopes, validOIDCClientScopesBearerAuthz) {
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidEntriesScope, config.Clients[c].ID, attrOIDCScopes, utils.StringJoinAnd(validOIDCClientScopesBearerAuthz), oidc.ScopeAutheliaBearerAuthz, utils.StringJoinAnd(config.Clients[c].Scopes)))
	}

	if len(config.Clients[c].GrantTypes) == 0 {
		validator.Push(fmt.Errorf(errFmtOIDCClientEmptyEntriesScope, config.Clients[c].ID, attrOIDCGrantTypes, utils.StringJoinAnd(validOIDCClientGrantTypesBearerAuthz), oidc.ScopeAutheliaBearerAuthz))
	} else {
		invalid, _ := validateList(config.Clients[c].GrantTypes, validOIDCClientGrantTypesBearerAuthz, false)

		if len(invalid) != 0 {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidEntriesScope, config.Clients[c].ID, attrOIDCGrantTypes, utils.StringJoinAnd(validOIDCClientGrantTypesBearerAuthz), oidc.ScopeAutheliaBearerAuthz, utils.StringJoinAnd(invalid)))
		}
	}

	if len(config.Clients[c].Audience) == 0 {
		validator.Push(fmt.Errorf(errFmtOIDCClientOptionRequiredScope, config.Clients[c].ID, "audience", oidc.ScopeAutheliaBearerAuthz))
	}

	if !ccg {
		if !config.Clients[c].RequirePushedAuthorizationRequests {
			validator.Push(fmt.Errorf(errFmtOIDCClientOptionMustScope, config.Clients[c].ID, "require_pushed_authorization_requests", "'true'", oidc.ScopeAutheliaBearerAuthz, "false"))
		}

		if !config.Clients[c].RequirePKCE {
			validator.Push(fmt.Errorf(errFmtOIDCClientOptionMustScope, config.Clients[c].ID, "require_pkce", "'true'", oidc.ScopeAutheliaBearerAuthz, "false"))
		} else if config.Clients[c].PKCEChallengeMethod != oidc.PKCEChallengeMethodSHA256 {
			validator.Push(fmt.Errorf(errFmtOIDCClientOptionMustScope, config.Clients[c].ID, attrOIDCPKCEChallengeMethod, "'"+oidc.PKCEChallengeMethodSHA256+"'", oidc.ScopeAutheliaBearerAuthz, config.Clients[c].PKCEChallengeMethod))
		}

		if config.Clients[c].ConsentMode != oidc.ClientConsentModeExplicit.String() {
			validator.Push(fmt.Errorf(errFmtOIDCClientOptionMustScope, config.Clients[c].ID, "consent_mode", "'"+oidc.ClientConsentModeExplicit.String()+"'", oidc.ScopeAutheliaBearerAuthz, config.Clients[c].ConsentMode))
		}

		if len(config.Clients[c].ResponseTypes) == 0 {
			validator.Push(fmt.Errorf(errFmtOIDCClientEmptyEntriesScope, config.Clients[c].ID, attrOIDCResponseTypes, utils.StringJoinAnd(validOIDCClientResponseTypesBearerAuthz), oidc.ScopeAutheliaBearerAuthz))
		} else if !utils.IsStringSliceContainsAll(config.Clients[c].ResponseTypes, validOIDCClientResponseTypesBearerAuthz) ||
			!utils.IsStringSliceContainsAny(config.Clients[c].ResponseTypes, validOIDCClientResponseTypesBearerAuthz) {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidEntriesScope, config.Clients[c].ID, attrOIDCResponseTypes, utils.StringJoinAnd(validOIDCClientResponseTypesBearerAuthz), oidc.ScopeAutheliaBearerAuthz, utils.StringJoinAnd(config.Clients[c].ResponseTypes)))
		}

		if len(config.Clients[c].ResponseModes) == 0 {
			validator.Push(fmt.Errorf(errFmtOIDCClientEmptyEntriesScope, config.Clients[c].ID, attrOIDCResponseModes, utils.StringJoinAnd(validOIDCClientResponseModesBearerAuthz), oidc.ScopeAutheliaBearerAuthz))
		} else if !utils.IsStringSliceContainsAll(config.Clients[c].ResponseModes, validOIDCClientResponseModesBearerAuthz) ||
			!utils.IsStringSliceContainsAny(config.Clients[c].ResponseModes, validOIDCClientResponseModesBearerAuthz) {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidEntriesScope, config.Clients[c].ID, attrOIDCResponseModes, utils.StringJoinAnd(validOIDCClientResponseModesBearerAuthz), oidc.ScopeAutheliaBearerAuthz, utils.StringJoinAnd(config.Clients[c].ResponseModes)))
		}
	}

	if config.Clients[c].Public {
		if config.Clients[c].TokenEndpointAuthMethod != oidc.ClientAuthMethodNone {
			validator.Push(fmt.Errorf(errFmtOIDCClientOptionMustScopeClientType, config.Clients[c].ID, attrOIDCTokenAuthMethod, "'"+oidc.ClientAuthMethodNone+"'", oidc.ScopeAutheliaBearerAuthz, "public", config.Clients[c].TokenEndpointAuthMethod))
		}
	} else {
		switch config.Clients[c].TokenEndpointAuthMethod {
		case oidc.ClientAuthMethodClientSecretBasic, oidc.ClientAuthMethodClientSecretJWT, oidc.ClientAuthMethodPrivateKeyJWT:
			break
		default:
			validator.Push(fmt.Errorf(errFmtOIDCClientOptionMustScopeClientType, config.Clients[c].ID, attrOIDCTokenAuthMethod, utils.StringJoinOr([]string{oidc.ClientAuthMethodClientSecretBasic, oidc.ClientAuthMethodClientSecretJWT, oidc.ClientAuthMethodPrivateKeyJWT}), oidc.ScopeAutheliaBearerAuthz, "confidential", config.Clients[c].TokenEndpointAuthMethod))
		}
	}

	return false
}

func validateOIDCClientScopesClientCredentialsGrant(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	invalid := validateListNotAllowed(config.Clients[c].Scopes, []string{oidc.ScopeOpenID, oidc.ScopeOffline, oidc.ScopeOfflineAccess})

	if len(invalid) > 0 {
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidEntriesClientCredentials, config.Clients[c].ID, utils.StringJoinAnd(config.Clients[c].Scopes), utils.StringJoinOr(invalid)))
	}
}

func validateOIDCClientResponseTypes(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator, setDefaults bool, errDeprecatedFunc func()) {
	if len(config.Clients[c].ResponseTypes) == 0 {
		if !setDefaults {
			return
		}

		config.Clients[c].ResponseTypes = schema.DefaultOpenIDConnectClientConfiguration.ResponseTypes
	}

	invalid, duplicates := validateList(config.Clients[c].ResponseTypes, validOIDCClientResponseTypes, true)

	if len(invalid) != 0 {
		validator.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidEntries, config.Clients[c].ID, attrOIDCResponseTypes, utils.StringJoinOr(validOIDCClientResponseTypes), utils.StringJoinAnd(invalid)))
	}

	if len(duplicates) != 0 {
		errDeprecatedFunc()

		validator.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidEntryDuplicates, config.Clients[c].ID, attrOIDCResponseTypes, utils.StringJoinAnd(duplicates)))
	}
}

func validateOIDCClientResponseModes(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator, setDefaults bool, errDeprecatedFunc func()) {
	if len(config.Clients[c].ResponseModes) == 0 {
		if !setDefaults {
			return
		}

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
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidEntries, config.Clients[c].ID, attrOIDCResponseModes, utils.StringJoinOr(validOIDCClientResponseModes), utils.StringJoinAnd(invalid)))
	}

	if len(duplicates) != 0 {
		errDeprecatedFunc()

		validator.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidEntryDuplicates, config.Clients[c].ID, attrOIDCResponseModes, utils.StringJoinAnd(duplicates)))
	}
}

func validateOIDCClientGrantTypes(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator, setDefaults bool, errDeprecatedFunc func()) {
	if len(config.Clients[c].GrantTypes) == 0 {
		if !setDefaults {
			return
		}

		validateOIDCClientGrantTypesSetDefaults(c, config)
	}

	validateOIDCClientGrantTypesCheckRelated(c, config, validator, errDeprecatedFunc)

	invalid, duplicates := validateList(config.Clients[c].GrantTypes, validOIDCClientGrantTypes, true)

	if len(invalid) != 0 {
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidEntries, config.Clients[c].ID, attrOIDCGrantTypes, utils.StringJoinOr(validOIDCClientGrantTypes), utils.StringJoinAnd(invalid)))
	}

	if len(duplicates) != 0 {
		errDeprecatedFunc()

		validator.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidEntryDuplicates, config.Clients[c].ID, attrOIDCGrantTypes, utils.StringJoinAnd(duplicates)))
	}
}

func validateOIDCClientGrantTypesSetDefaults(c int, config *schema.IdentityProvidersOpenIDConnect) {
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

func validateOIDCClientGrantTypesCheckRelated(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator, errDeprecatedFunc func()) {
	for _, grantType := range config.Clients[c].GrantTypes {
		switch grantType {
		case oidc.GrantTypeAuthorizationCode:
			if !utils.IsStringInSlice(oidc.ResponseTypeAuthorizationCodeFlow, config.Clients[c].ResponseTypes) && !utils.IsStringSliceContainsAny(validOIDCClientResponseTypesHybridFlow, config.Clients[c].ResponseTypes) {
				errDeprecatedFunc()

				validator.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidGrantTypeMatch, config.Clients[c].ID, grantType, "for either the authorization code or hybrid flow", utils.StringJoinOr(append([]string{oidc.ResponseTypeAuthorizationCodeFlow}, validOIDCClientResponseTypesHybridFlow...)), utils.StringJoinAnd(config.Clients[c].ResponseTypes)))
			}
		case oidc.GrantTypeImplicit:
			if !utils.IsStringSliceContainsAny(validOIDCClientResponseTypesImplicitFlow, config.Clients[c].ResponseTypes) && !utils.IsStringSliceContainsAny(validOIDCClientResponseTypesHybridFlow, config.Clients[c].ResponseTypes) {
				errDeprecatedFunc()

				validator.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidGrantTypeMatch, config.Clients[c].ID, grantType, "for either the implicit or hybrid flow", utils.StringJoinOr(append(append([]string{}, validOIDCClientResponseTypesImplicitFlow...), validOIDCClientResponseTypesHybridFlow...)), utils.StringJoinAnd(config.Clients[c].ResponseTypes)))
			}
		case oidc.GrantTypeClientCredentials:
			if config.Clients[c].Public {
				validator.Push(fmt.Errorf(errFmtOIDCClientInvalidGrantTypePublic, config.Clients[c].ID, oidc.GrantTypeClientCredentials))
			}
		case oidc.GrantTypeRefreshToken:
			if !utils.IsStringSliceContainsAny([]string{oidc.ScopeOfflineAccess, oidc.ScopeOffline}, config.Clients[c].Scopes) {
				errDeprecatedFunc()

				validator.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidGrantTypeRefresh, config.Clients[c].ID))
			}

			if !utils.IsStringSliceContainsAny(validOIDCClientResponseTypesRefreshToken, config.Clients[c].ResponseTypes) {
				errDeprecatedFunc()

				validator.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidRefreshTokenOptionWithoutCodeResponseType,
					config.Clients[c].ID, attrOIDCGrantTypes,
					utils.StringJoinOr([]string{oidc.GrantTypeRefreshToken}),
					utils.StringJoinOr(validOIDCClientResponseTypesRefreshToken)),
				)
			}
		}
	}
}

func validateOIDCClientRedirectURIs(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator, errDeprecatedFunc func()) {
	var (
		parsedRedirectURI *url.URL
		err               error
	)

	for _, redirectURI := range config.Clients[c].RedirectURIs {
		if redirectURI == oidc.RedirectURISpecialOAuth2InstalledApp {
			if config.Clients[c].Public {
				continue
			}

			validator.Push(fmt.Errorf(errFmtOIDCClientRedirectURIPublic, config.Clients[c].ID, oidc.RedirectURISpecialOAuth2InstalledApp))

			continue
		}

		if parsedRedirectURI, err = url.Parse(redirectURI); err != nil {
			validator.Push(fmt.Errorf(errFmtOIDCClientRedirectURICantBeParsed, config.Clients[c].ID, redirectURI, err))
			continue
		}

		if !parsedRedirectURI.IsAbs() || (!config.Clients[c].Public && parsedRedirectURI.Scheme == "") {
			validator.Push(fmt.Errorf(errFmtOIDCClientRedirectURIAbsolute, config.Clients[c].ID, redirectURI))
		}
	}

	_, duplicates := validateList(config.Clients[c].RedirectURIs, nil, true)

	if len(duplicates) != 0 {
		errDeprecatedFunc()

		validator.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidEntryDuplicates, config.Clients[c].ID, attrOIDCRedirectURIs, utils.StringJoinAnd(duplicates)))
	}
}

func validateOIDCClientRequestURIs(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	var (
		parsedRequestURI *url.URL
		err              error
	)

	for _, requestURI := range config.Clients[c].RequestURIs {
		if parsedRequestURI, err = url.Parse(requestURI); err != nil {
			validator.Push(fmt.Errorf(errFmtOIDCClientRequestURICantBeParsed, config.Clients[c].ID, requestURI, err))
			continue
		}

		if !parsedRequestURI.IsAbs() {
			validator.Push(fmt.Errorf(errFmtOIDCClientRequestURINotAbsolute, config.Clients[c].ID, requestURI))
		} else if parsedRequestURI.Scheme != schemeHTTPS {
			validator.Push(fmt.Errorf(errFmtOIDCClientRequestURIInvalidScheme, config.Clients[c].ID, requestURI, parsedRequestURI.Scheme))
		}
	}

	_, duplicates := validateList(config.Clients[c].RequestURIs, nil, true)

	if len(duplicates) != 0 {
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidEntryDuplicates, config.Clients[c].ID, attrOIDCRequestURIs, utils.StringJoinAnd(duplicates)))
	}
}

//nolint:gocyclo
func validateOIDCClientTokenEndpointAuth(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	implicit := len(config.Clients[c].ResponseTypes) != 0 && utils.IsStringSliceContainsAll(config.Clients[c].ResponseTypes, validOIDCClientResponseTypesImplicitFlow)

	switch {
	case config.Clients[c].TokenEndpointAuthMethod == "":
		if config.Clients[c].Public {
			config.Clients[c].TokenEndpointAuthMethod = oidc.ClientAuthMethodNone
		} else {
			config.Clients[c].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretBasic
		}
	case !utils.IsStringInSlice(config.Clients[c].TokenEndpointAuthMethod, validOIDCClientTokenEndpointAuthMethods):
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
			config.Clients[c].ID, attrOIDCTokenAuthMethod, utils.StringJoinOr(validOIDCClientTokenEndpointAuthMethods), config.Clients[c].TokenEndpointAuthMethod))

		return
	case config.Clients[c].TokenEndpointAuthMethod == oidc.ClientAuthMethodNone && !config.Clients[c].Public && !implicit:
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidTokenEndpointAuthMethod,
			config.Clients[c].ID, utils.StringJoinOr(validOIDCClientTokenEndpointAuthMethodsConfidential), utils.StringJoinAnd(validOIDCClientResponseTypesImplicitFlow), config.Clients[c].TokenEndpointAuthMethod))
	case config.Clients[c].TokenEndpointAuthMethod != oidc.ClientAuthMethodNone && config.Clients[c].Public:
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidTokenEndpointAuthMethodPublic,
			config.Clients[c].ID, config.Clients[c].TokenEndpointAuthMethod))
	}

	secret := false

	switch config.Clients[c].TokenEndpointAuthMethod {
	case oidc.ClientAuthMethodClientSecretJWT:
		validateOIDCClientTokenEndpointAuthClientSecretJWT(c, config, validator)

		secret = true
	case "":
		if !config.Clients[c].Public {
			secret = true
		}
	case oidc.ClientAuthMethodClientSecretPost, oidc.ClientAuthMethodClientSecretBasic:
		secret = true
	case oidc.ClientAuthMethodPrivateKeyJWT:
		validateOIDCClientTokenEndpointAuthPublicKeyJWT(config.Clients[c], validator)
	}

	if secret {
		if config.Clients[c].Public {
			return
		}

		if !config.Clients[c].Secret.Valid() {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidSecret, config.Clients[c].ID))
		} else {
			switch {
			case config.Clients[c].Secret.IsPlainText() && config.Clients[c].TokenEndpointAuthMethod != oidc.ClientAuthMethodClientSecretJWT:
				validator.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidSecretPlainText, config.Clients[c].ID))
			case !config.Clients[c].Secret.IsPlainText() && config.Clients[c].TokenEndpointAuthMethod == oidc.ClientAuthMethodClientSecretJWT:
				validator.Push(fmt.Errorf(errFmtOIDCClientInvalidSecretNotPlainText, config.Clients[c].ID))
			}
		}
	} else if config.Clients[c].Secret != nil {
		if config.Clients[c].Public {
			validator.Push(fmt.Errorf(errFmtOIDCClientPublicInvalidSecret, config.Clients[c].ID))
		} else {
			validator.Push(fmt.Errorf(errFmtOIDCClientPublicInvalidSecretClientAuthMethod, config.Clients[c].ID, config.Clients[c].TokenEndpointAuthMethod))
		}
	}
}

func validateOIDCClientTokenEndpointAuthClientSecretJWT(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	switch {
	case config.Clients[c].TokenEndpointAuthSigningAlg == "":
		config.Clients[c].TokenEndpointAuthSigningAlg = oidc.SigningAlgHMACUsingSHA256
	case !utils.IsStringInSlice(config.Clients[c].TokenEndpointAuthSigningAlg, validOIDCClientTokenEndpointAuthSigAlgsClientSecretJWT):
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidTokenEndpointAuthSigAlg, config.Clients[c].ID, utils.StringJoinOr(validOIDCClientTokenEndpointAuthSigAlgsClientSecretJWT), config.Clients[c].TokenEndpointAuthMethod))
	}
}

func validateOIDCClientTokenEndpointAuthPublicKeyJWT(config schema.IdentityProvidersOpenIDConnectClient, validator *schema.StructValidator) {
	switch {
	case config.TokenEndpointAuthSigningAlg == "":
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidTokenEndpointAuthSigAlgMissingPrivateKeyJWT, config.ID))
	case !utils.IsStringInSlice(config.TokenEndpointAuthSigningAlg, validOIDCIssuerJWKSigningAlgs):
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidTokenEndpointAuthSigAlg, config.ID, utils.StringJoinOr(validOIDCIssuerJWKSigningAlgs), config.TokenEndpointAuthMethod))
	}

	if config.JSONWebKeysURI == nil {
		if len(config.JSONWebKeys) == 0 {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidPublicKeysPrivateKeyJWT, config.ID))
		} else if len(config.Discovery.RequestObjectSigningAlgs) != 0 && !utils.IsStringInSlice(config.TokenEndpointAuthSigningAlg, config.Discovery.RequestObjectSigningAlgs) {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidTokenEndpointAuthSigAlgReg, config.ID, utils.StringJoinOr(config.Discovery.RequestObjectSigningAlgs), config.TokenEndpointAuthMethod))
		}
	}
}

func validateOIDDClientSigningAlgs(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	validateOIDDClientSigningAlgsJARM(c, config, validator)
	validateOIDDClientSigningAlgsIDToken(c, config, validator)
	validateOIDDClientSigningAlgsAccessToken(c, config, validator)
	validateOIDDClientSigningAlgsUserInfo(c, config, validator)
	validateOIDDClientSigningAlgsIntrospection(c, config, validator)
}

func validateOIDDClientSigningAlgsIDToken(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	config.Clients[c].IDTokenSignedResponseAlg, config.Clients[c].IDTokenSignedResponseKeyID = validateOIDCAlgKIDDefault(config, config.Clients[c].IDTokenSignedResponseAlg, config.Clients[c].IDTokenSignedResponseKeyID, schema.DefaultOpenIDConnectClientConfiguration.IDTokenSignedResponseAlg)

	switch config.Clients[c].IDTokenSignedResponseKeyID {
	case "":
		switch config.Clients[c].IDTokenSignedResponseAlg {
		case "", oidc.SigningAlgRSAUsingSHA256:
			break
		default:
			if !utils.IsStringInSlice(config.Clients[c].IDTokenSignedResponseAlg, config.Discovery.ResponseObjectSigningAlgs) {
				validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
					config.Clients[c].ID, attrOIDCIDTokenSigAlg, utils.StringJoinOr(config.Discovery.ResponseObjectSigningAlgs), config.Clients[c].IDTokenSignedResponseAlg))
			}
		}
	default:
		if !utils.IsStringInSlice(config.Clients[c].IDTokenSignedResponseKeyID, config.Discovery.ResponseObjectSigningKeyIDs) {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
				config.Clients[c].ID, attrOIDCIDTokenSigKID, utils.StringJoinOr(config.Discovery.ResponseObjectSigningKeyIDs), config.Clients[c].IDTokenSignedResponseKeyID))
		} else {
			config.Clients[c].IDTokenSignedResponseAlg = getResponseObjectAlgFromKID(config, config.Clients[c].IDTokenSignedResponseKeyID, config.Clients[c].IDTokenSignedResponseAlg)
		}
	}
}

func validateOIDDClientSigningAlgsAccessToken(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	config.Clients[c].AccessTokenSignedResponseAlg, config.Clients[c].AccessTokenSignedResponseKeyID = validateOIDCAlgKIDDefault(config, config.Clients[c].AccessTokenSignedResponseAlg, config.Clients[c].AccessTokenSignedResponseKeyID, schema.DefaultOpenIDConnectClientConfiguration.AccessTokenSignedResponseAlg)

	switch config.Clients[c].AccessTokenSignedResponseKeyID {
	case "":
		switch config.Clients[c].AccessTokenSignedResponseAlg {
		case "", oidc.SigningAlgNone, oidc.SigningAlgRSAUsingSHA256:
			break
		default:
			if !utils.IsStringInSlice(config.Clients[c].AccessTokenSignedResponseAlg, config.Discovery.ResponseObjectSigningAlgs) {
				validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
					config.Clients[c].ID, attrOIDCAccessTokenSigAlg, utils.StringJoinOr(config.Discovery.ResponseObjectSigningAlgs), config.Clients[c].AccessTokenSignedResponseAlg))
			} else {
				config.Discovery.JWTResponseAccessTokens = true
			}
		}
	default:
		switch {
		case !utils.IsStringInSlice(config.Clients[c].AccessTokenSignedResponseKeyID, config.Discovery.ResponseObjectSigningKeyIDs):
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
				config.Clients[c].ID, attrOIDCAccessTokenSigKID, utils.StringJoinOr(config.Discovery.ResponseObjectSigningKeyIDs), config.Clients[c].AccessTokenSignedResponseKeyID))
		default:
			config.Clients[c].AccessTokenSignedResponseAlg = getResponseObjectAlgFromKID(config, config.Clients[c].AccessTokenSignedResponseKeyID, config.Clients[c].AccessTokenSignedResponseAlg)
			config.Discovery.JWTResponseAccessTokens = true
		}
	}
}

func validateOIDDClientSigningAlgsUserInfo(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	config.Clients[c].UserinfoSignedResponseAlg, config.Clients[c].UserinfoSignedResponseKeyID = validateOIDCAlgKIDDefault(config, config.Clients[c].UserinfoSignedResponseAlg, config.Clients[c].UserinfoSignedResponseKeyID, schema.DefaultOpenIDConnectClientConfiguration.UserinfoSignedResponseAlg)

	switch config.Clients[c].UserinfoSignedResponseKeyID {
	case "":
		switch config.Clients[c].UserinfoSignedResponseAlg {
		case "", oidc.SigningAlgNone, oidc.SigningAlgRSAUsingSHA256:
			break
		default:
			if !utils.IsStringInSlice(config.Clients[c].UserinfoSignedResponseAlg, config.Discovery.ResponseObjectSigningAlgs) {
				validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
					config.Clients[c].ID, attrOIDCUsrSigAlg, utils.StringJoinOr(append(config.Discovery.ResponseObjectSigningAlgs, oidc.SigningAlgNone)), config.Clients[c].UserinfoSignedResponseAlg))
			}
		}
	default:
		if !utils.IsStringInSlice(config.Clients[c].UserinfoSignedResponseKeyID, config.Discovery.ResponseObjectSigningKeyIDs) {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
				config.Clients[c].ID, attrOIDCUsrSigKID, utils.StringJoinOr(config.Discovery.ResponseObjectSigningKeyIDs), config.Clients[c].UserinfoSignedResponseKeyID))
		} else {
			config.Clients[c].UserinfoSignedResponseAlg = getResponseObjectAlgFromKID(config, config.Clients[c].UserinfoSignedResponseKeyID, config.Clients[c].UserinfoSignedResponseAlg)
		}
	}
}

func validateOIDDClientSigningAlgsIntrospection(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	config.Clients[c].IntrospectionSignedResponseAlg, config.Clients[c].IntrospectionSignedResponseKeyID = validateOIDCAlgKIDDefault(config, config.Clients[c].IntrospectionSignedResponseAlg, config.Clients[c].IntrospectionSignedResponseKeyID, schema.DefaultOpenIDConnectClientConfiguration.IntrospectionSignedResponseAlg)

	switch config.Clients[c].IntrospectionSignedResponseKeyID {
	case "":
		switch config.Clients[c].IntrospectionSignedResponseAlg {
		case "", oidc.SigningAlgNone, oidc.SigningAlgRSAUsingSHA256:
			break
		default:
			if !utils.IsStringInSlice(config.Clients[c].IntrospectionSignedResponseAlg, config.Discovery.ResponseObjectSigningAlgs) {
				validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
					config.Clients[c].ID, attrOIDCIntrospectionSigAlg, utils.StringJoinOr(append(config.Discovery.ResponseObjectSigningAlgs, oidc.SigningAlgNone)), config.Clients[c].IntrospectionSignedResponseAlg))
			}
		}
	default:
		if !utils.IsStringInSlice(config.Clients[c].IntrospectionSignedResponseKeyID, config.Discovery.ResponseObjectSigningKeyIDs) {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
				config.Clients[c].ID, attrOIDCIntrospectionSigKID, utils.StringJoinOr(config.Discovery.ResponseObjectSigningKeyIDs), config.Clients[c].IntrospectionSignedResponseKeyID))
		} else {
			config.Clients[c].IntrospectionSignedResponseAlg = getResponseObjectAlgFromKID(config, config.Clients[c].IntrospectionSignedResponseKeyID, config.Clients[c].IntrospectionSignedResponseAlg)
		}
	}
}

func validateOIDDClientSigningAlgsJARM(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	config.Clients[c].AuthorizationSignedResponseAlg, config.Clients[c].AuthorizationSignedResponseKeyID = validateOIDCAlgKIDDefault(config, config.Clients[c].AuthorizationSignedResponseAlg, config.Clients[c].AuthorizationSignedResponseKeyID, schema.DefaultOpenIDConnectClientConfiguration.AuthorizationSignedResponseAlg)

	switch config.Clients[c].AuthorizationSignedResponseKeyID {
	case "":
		switch config.Clients[c].AuthorizationSignedResponseAlg {
		case "", oidc.SigningAlgNone, oidc.SigningAlgRSAUsingSHA256:
			break
		default:
			if !utils.IsStringInSlice(config.Clients[c].AuthorizationSignedResponseAlg, config.Discovery.ResponseObjectSigningAlgs) {
				validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
					config.Clients[c].ID, attrOIDCAuthorizationSigAlg, utils.StringJoinOr(config.Discovery.ResponseObjectSigningAlgs), config.Clients[c].AuthorizationSignedResponseAlg))
			}
		}
	default:
		if !utils.IsStringInSlice(config.Clients[c].AuthorizationSignedResponseKeyID, config.Discovery.ResponseObjectSigningKeyIDs) {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
				config.Clients[c].ID, attrOIDCAuthorizationSigKID, utils.StringJoinOr(config.Discovery.ResponseObjectSigningKeyIDs), config.Clients[c].AuthorizationSignedResponseKeyID))
		} else {
			config.Clients[c].AuthorizationSignedResponseAlg = getResponseObjectAlgFromKID(config, config.Clients[c].AuthorizationSignedResponseKeyID, config.Clients[c].AuthorizationSignedResponseAlg)
		}
	}
}

func validateOIDCAlgKIDDefault(config *schema.IdentityProvidersOpenIDConnect, algCurrent, kidCurrent, algDefault string) (alg, kid string) {
	alg, kid = algCurrent, kidCurrent

	switch balg, bkid := len(alg) != 0, len(kid) != 0; {
	case balg && bkid:
		return
	case !balg && !bkid:
		if algDefault == "" {
			return
		}

		alg = algDefault
	}

	switch balg, bkid := len(alg) != 0, len(kid) != 0; {
	case !balg && !bkid:
		return
	case !bkid:
		for _, jwk := range config.JSONWebKeys {
			if alg == jwk.Algorithm {
				kid = jwk.KeyID

				return
			}
		}
	case !balg:
		for _, jwk := range config.JSONWebKeys {
			if kid == jwk.KeyID {
				alg = jwk.Algorithm

				return
			}
		}
	}

	return
}
