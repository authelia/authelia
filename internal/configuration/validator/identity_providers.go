package validator

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateIdentityProviders validates and updates the IdentityProviders configuration.
func ValidateIdentityProviders(config *schema.IdentityProvidersConfiguration, val *schema.StructValidator) {
	validateOIDC(config.OIDC, val)
}

func validateOIDC(config *schema.OpenIDConnectConfiguration, val *schema.StructValidator) {
	if config == nil {
		return
	}

	setOIDCDefaults(config)

	switch {
	case config.IssuerPrivateKey == nil:
		val.Push(fmt.Errorf(errFmtOIDCNoPrivateKey))
	default:
		if config.IssuerCertificateChain.HasCertificates() {
			if !config.IssuerCertificateChain.EqualKey(config.IssuerPrivateKey) {
				val.Push(fmt.Errorf(errFmtOIDCCertificateMismatch))
			}

			if err := config.IssuerCertificateChain.Validate(); err != nil {
				val.Push(fmt.Errorf(errFmtOIDCCertificateChain, err))
			}
		}

		if config.IssuerPrivateKey.Size()*8 < 2048 {
			val.Push(fmt.Errorf(errFmtOIDCInvalidPrivateKeyBitSize, 2048, config.IssuerPrivateKey.Size()*8))
		}
	}

	if config.MinimumParameterEntropy != 0 && config.MinimumParameterEntropy < 8 {
		val.PushWarning(fmt.Errorf(errFmtOIDCServerInsecureParameterEntropy, config.MinimumParameterEntropy))
	}

	if config.EnforcePKCE != "never" && config.EnforcePKCE != "public_clients_only" && config.EnforcePKCE != "always" {
		val.Push(fmt.Errorf(errFmtOIDCEnforcePKCEInvalidValue, config.EnforcePKCE))
	}

	validateOIDCOptionsCORS(config, val)

	if len(config.Clients) == 0 {
		val.Push(fmt.Errorf(errFmtOIDCNoClientsConfigured))
	} else {
		validateOIDCClients(config, val)
	}
}

func setOIDCDefaults(config *schema.OpenIDConnectConfiguration) {
	if config.AccessTokenLifespan == time.Duration(0) {
		config.AccessTokenLifespan = schema.DefaultOpenIDConnectConfiguration.AccessTokenLifespan
	}

	if config.AuthorizeCodeLifespan == time.Duration(0) {
		config.AuthorizeCodeLifespan = schema.DefaultOpenIDConnectConfiguration.AuthorizeCodeLifespan
	}

	if config.IDTokenLifespan == time.Duration(0) {
		config.IDTokenLifespan = schema.DefaultOpenIDConnectConfiguration.IDTokenLifespan
	}

	if config.RefreshTokenLifespan == time.Duration(0) {
		config.RefreshTokenLifespan = schema.DefaultOpenIDConnectConfiguration.RefreshTokenLifespan
	}

	if config.EnforcePKCE == "" {
		config.EnforcePKCE = schema.DefaultOpenIDConnectConfiguration.EnforcePKCE
	}
}

func validateOIDCOptionsCORS(config *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	validateOIDCOptionsCORSAllowedOrigins(config, validator)

	if config.CORS.AllowedOriginsFromClientRedirectURIs {
		validateOIDCOptionsCORSAllowedOriginsFromClientRedirectURIs(config)
	}

	validateOIDCOptionsCORSEndpoints(config, validator)
}

func validateOIDCOptionsCORSAllowedOrigins(config *schema.OpenIDConnectConfiguration, val *schema.StructValidator) {
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

func validateOIDCOptionsCORSAllowedOriginsFromClientRedirectURIs(config *schema.OpenIDConnectConfiguration) {
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

func validateOIDCOptionsCORSEndpoints(config *schema.OpenIDConnectConfiguration, val *schema.StructValidator) {
	for _, endpoint := range config.CORS.Endpoints {
		if !utils.IsStringInSlice(endpoint, validOIDCCORSEndpoints) {
			val.Push(fmt.Errorf(errFmtOIDCCORSInvalidEndpoint, endpoint, strJoinOr(validOIDCCORSEndpoints)))
		}
	}
}

func validateOIDCClients(config *schema.OpenIDConnectConfiguration, val *schema.StructValidator) {
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

func validateOIDCClient(c int, config *schema.OpenIDConnectConfiguration, val *schema.StructValidator, errDeprecatedFunc func()) {
	if config.Clients[c].Public {
		if config.Clients[c].Secret != nil {
			val.Push(fmt.Errorf(errFmtOIDCClientPublicInvalidSecret, config.Clients[c].ID))
		}
	} else {
		if config.Clients[c].Secret == nil {
			val.Push(fmt.Errorf(errFmtOIDCClientInvalidSecret, config.Clients[c].ID))
		} else if config.Clients[c].Secret.IsPlainText() {
			val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidSecretPlainText, config.Clients[c].ID))
		}
	}

	switch config.Clients[c].Policy {
	case "":
		config.Clients[c].Policy = schema.DefaultOpenIDConnectClientConfiguration.Policy
	case policyOneFactor, policyTwoFactor:
		break
	default:
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidValue, config.Clients[c].ID, "policy", strJoinOr([]string{policyOneFactor, policyTwoFactor}), config.Clients[c].Policy))
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

	validateOIDCClientTokenEndpointAuthMethod(c, config, val)
	validateOIDDClientUserinfoAlgorithm(c, config, val)

	validateOIDCClientSectorIdentifier(c, config, val)
}

func validateOIDCClientSectorIdentifier(c int, config *schema.OpenIDConnectConfiguration, val *schema.StructValidator) {
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

func validateOIDCClientConsentMode(c int, config *schema.OpenIDConnectConfiguration, val *schema.StructValidator) {
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

func validateOIDCClientScopes(c int, config *schema.OpenIDConnectConfiguration, val *schema.StructValidator, errDeprecatedFunc func()) {
	if len(config.Clients[c].Scopes) == 0 {
		config.Clients[c].Scopes = schema.DefaultOpenIDConnectClientConfiguration.Scopes
	}

	if !utils.IsStringInSlice(oidc.ScopeOpenID, config.Clients[c].Scopes) {
		config.Clients[c].Scopes = append([]string{oidc.ScopeOpenID}, config.Clients[c].Scopes...)
	}

	invalid, duplicates := validateList(config.Clients[c].Scopes, validOIDCClientScopes, true)

	if len(invalid) != 0 {
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidEntries, config.Clients[c].ID, attrOIDCScopes, strJoinOr(validOIDCClientScopes), strJoinAnd(invalid)))
	}

	if len(duplicates) != 0 {
		errDeprecatedFunc()

		val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidEntryDuplicates, config.Clients[c].ID, attrOIDCScopes, strJoinAnd(duplicates)))
	}
}

func validateOIDCClientResponseTypes(c int, config *schema.OpenIDConnectConfiguration, val *schema.StructValidator, errDeprecatedFunc func()) {
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

func validateOIDCClientResponseModes(c int, config *schema.OpenIDConnectConfiguration, val *schema.StructValidator, errDeprecatedFunc func()) {
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

func validateOIDCClientGrantTypes(c int, config *schema.OpenIDConnectConfiguration, val *schema.StructValidator, errDeprecatedFunc func()) {
	if len(config.Clients[c].GrantTypes) == 0 {
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

	for _, grantType := range config.Clients[c].GrantTypes {
		switch grantType {
		case oidc.GrantTypeImplicit:
			if !utils.IsStringSliceContainsAny(validOIDCClientResponseModesImplicitFlow, config.Clients[c].ResponseTypes) && !utils.IsStringSliceContainsAny(validOIDCClientResponseModesHybridFlow, config.Clients[c].ResponseTypes) {
				errDeprecatedFunc()

				val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidGrantTypeMatch, config.Clients[c].ID, grantType, "for either the implicit or hybrid flow", strJoinOr(append(append([]string{}, validOIDCClientResponseModesImplicitFlow...), validOIDCClientResponseModesHybridFlow...)), strJoinAnd(config.Clients[c].ResponseTypes)))
			}
		case oidc.GrantTypeAuthorizationCode:
			if !utils.IsStringInSlice(oidc.ResponseTypeAuthorizationCodeFlow, config.Clients[c].ResponseTypes) && !utils.IsStringSliceContainsAny(validOIDCClientResponseModesHybridFlow, config.Clients[c].ResponseTypes) {
				errDeprecatedFunc()

				val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidGrantTypeMatch, config.Clients[c].ID, grantType, "for either the authorization code or hybrid flow", strJoinOr(append([]string{oidc.ResponseTypeAuthorizationCodeFlow}, validOIDCClientResponseModesHybridFlow...)), strJoinAnd(config.Clients[c].ResponseTypes)))
			}
		case oidc.GrantTypeRefreshToken:
			if !utils.IsStringInSlice(oidc.ScopeOfflineAccess, config.Clients[c].Scopes) {
				errDeprecatedFunc()

				val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidGrantTypeRefresh, config.Clients[c].ID))
			}
		}
	}

	invalid, duplicates := validateList(config.Clients[c].GrantTypes, validOIDCClientGrantTypes, true)

	if len(invalid) != 0 {
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidEntries, config.Clients[c].ID, attrOIDCGrantTypes, strJoinOr(validOIDCClientGrantTypes), strJoinAnd(invalid)))
	}

	if len(duplicates) != 0 {
		errDeprecatedFunc()

		val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidEntryDuplicates, config.Clients[c].ID, attrOIDCGrantTypes, strJoinAnd(duplicates)))
	}
}

func validateOIDCClientRedirectURIs(c int, config *schema.OpenIDConnectConfiguration, val *schema.StructValidator, errDeprecatedFunc func()) {
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
			return
		}
	}

	_, duplicates := validateList(config.Clients[c].RedirectURIs, nil, true)

	if len(duplicates) != 0 {
		errDeprecatedFunc()

		val.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidEntryDuplicates, config.Clients[c].ID, attrOIDCRedirectURIs, strJoinAnd(duplicates)))
	}
}

func validateOIDCClientTokenEndpointAuthMethod(c int, config *schema.OpenIDConnectConfiguration, val *schema.StructValidator) {
	if config.Clients[c].TokenEndpointAuthMethod == "" {
		if config.Clients[c].Public {
			config.Clients[c].TokenEndpointAuthMethod = oidc.ClientAuthMethodNone
		} else {
			config.Clients[c].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretBasic
		}
	}

	switch {
	case !utils.IsStringInSlice(config.Clients[c].TokenEndpointAuthMethod, validOIDCClientTokenEndpointAuthMethods):
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
			config.Clients[c].ID, attrOIDCTokenAuthMethod, strJoinOr(validOIDCClientTokenEndpointAuthMethods), config.Clients[c].TokenEndpointAuthMethod))
	case config.Clients[c].TokenEndpointAuthMethod == oidc.ClientAuthMethodNone && !config.Clients[c].Public:
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidTokenEndpointAuthMethod,
			config.Clients[c].ID, strJoinOr(validOIDCClientTokenEndpointAuthMethodsConfidential), attrOIDCConfidential, config.Clients[c].TokenEndpointAuthMethod))
	case config.Clients[c].TokenEndpointAuthMethod != oidc.ClientAuthMethodNone && config.Clients[c].Public:
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidTokenEndpointAuthMethod,
			config.Clients[c].ID, strJoinOr([]string{oidc.ClientAuthMethodNone}), attrOIDCPublic, config.Clients[c].TokenEndpointAuthMethod))
	}
}

func validateOIDDClientUserinfoAlgorithm(c int, config *schema.OpenIDConnectConfiguration, val *schema.StructValidator) {
	if config.Clients[c].UserinfoSigningAlgorithm == "" {
		config.Clients[c].UserinfoSigningAlgorithm = schema.DefaultOpenIDConnectClientConfiguration.UserinfoSigningAlgorithm
	}

	if !utils.IsStringInSlice(config.Clients[c].UserinfoSigningAlgorithm, validOIDCClientUserinfoAlgorithms) {
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
			config.Clients[c].ID, attrOIDCUsrSigAlg, strJoinOr(validOIDCClientUserinfoAlgorithms), config.Clients[c].UserinfoSigningAlgorithm))
	}
}
