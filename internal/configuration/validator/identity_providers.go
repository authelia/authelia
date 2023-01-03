package validator

import (
	"fmt"
	"net/url"
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

			origin := utils.OriginFromURL(*uri)

			if !utils.IsURLInSlice(origin, config.CORS.AllowedOrigins) {
				config.CORS.AllowedOrigins = append(config.CORS.AllowedOrigins, origin)
			}
		}
	}
}

func validateOIDCOptionsCORSEndpoints(config *schema.OpenIDConnectConfiguration, val *schema.StructValidator) {
	for _, endpoint := range config.CORS.Endpoints {
		if !utils.IsStringInSlice(endpoint, validOIDCCORSEndpoints) {
			val.Push(fmt.Errorf(errFmtOIDCCORSInvalidEndpoint, endpoint, strings.Join(validOIDCCORSEndpoints, "', '")))
		}
	}
}

func validateOIDCClients(config *schema.OpenIDConnectConfiguration, val *schema.StructValidator) {
	invalidID, duplicateIDs := false, false

	var ids []string

	for c, client := range config.Clients {
		if client.ID == "" {
			invalidID = true
		} else {
			if client.Description == "" {
				config.Clients[c].Description = client.ID
			}

			if utils.IsStringInSliceFold(client.ID, ids) {
				duplicateIDs = true
			}
			ids = append(ids, client.ID)
		}

		if client.Public {
			if client.Secret != nil {
				val.Push(fmt.Errorf(errFmtOIDCClientPublicInvalidSecret, client.ID))
			}
		} else {
			if client.Secret == nil {
				val.Push(fmt.Errorf(errFmtOIDCClientInvalidSecret, client.ID))
			}
		}

		if client.Policy == "" {
			config.Clients[c].Policy = schema.DefaultOpenIDConnectClientConfiguration.Policy
		} else if client.Policy != policyOneFactor && client.Policy != policyTwoFactor {
			val.Push(fmt.Errorf(errFmtOIDCClientInvalidPolicy, client.ID, client.Policy))
		}

		switch client.PKCEChallengeMethod {
		case "", "plain", "S256":
			break
		default:
			val.Push(fmt.Errorf(errFmtOIDCClientInvalidPKCEChallengeMethod, client.ID, client.PKCEChallengeMethod))
		}

		validateOIDCClientConsentMode(c, config, val)
		validateOIDCClientSectorIdentifier(client, val)
		validateOIDCClientScopes(c, config, val)
		validateOIDCClientGrantTypes(c, config, val)
		validateOIDCClientResponseTypes(c, config, val)
		validateOIDCClientResponseModes(c, config, val)
		validateOIDDClientUserinfoAlgorithm(c, config, val)
		validateOIDCClientRedirectURIs(client, val)
	}

	if invalidID {
		val.Push(fmt.Errorf(errFmtOIDCClientsWithEmptyID))
	}

	if duplicateIDs {
		val.Push(fmt.Errorf(errFmtOIDCClientsDuplicateID))
	}
}

func validateOIDCClientSectorIdentifier(client schema.OpenIDConnectClientConfiguration, val *schema.StructValidator) {
	if client.SectorIdentifier.String() != "" {
		if utils.IsURLHostComponent(client.SectorIdentifier) || utils.IsURLHostComponentWithPort(client.SectorIdentifier) {
			return
		}

		if client.SectorIdentifier.Scheme != "" {
			val.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifier, client.ID, client.SectorIdentifier.String(), client.SectorIdentifier.Host, "scheme", client.SectorIdentifier.Scheme))

			if client.SectorIdentifier.Path != "" {
				val.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifier, client.ID, client.SectorIdentifier.String(), client.SectorIdentifier.Host, "path", client.SectorIdentifier.Path))
			}

			if client.SectorIdentifier.RawQuery != "" {
				val.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifier, client.ID, client.SectorIdentifier.String(), client.SectorIdentifier.Host, "query", client.SectorIdentifier.RawQuery))
			}

			if client.SectorIdentifier.Fragment != "" {
				val.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifier, client.ID, client.SectorIdentifier.String(), client.SectorIdentifier.Host, "fragment", client.SectorIdentifier.Fragment))
			}

			if client.SectorIdentifier.User != nil {
				if client.SectorIdentifier.User.Username() != "" {
					val.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifier, client.ID, client.SectorIdentifier.String(), client.SectorIdentifier.Host, "username", client.SectorIdentifier.User.Username()))
				}

				if _, set := client.SectorIdentifier.User.Password(); set {
					val.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifierWithoutValue, client.ID, client.SectorIdentifier.String(), client.SectorIdentifier.Host, "password"))
				}
			}
		} else if client.SectorIdentifier.Host == "" {
			val.Push(fmt.Errorf(errFmtOIDCClientInvalidSectorIdentifierHost, client.ID, client.SectorIdentifier.String()))
		}
	}
}

func validateOIDCClientConsentMode(c int, config *schema.OpenIDConnectConfiguration, val *schema.StructValidator) {
	switch {
	case utils.IsStringInSlice(config.Clients[c].ConsentMode, []string{"", "auto"}):
		if config.Clients[c].ConsentPreConfiguredDuration != nil {
			config.Clients[c].ConsentMode = oidc.ClientConsentModePreConfigured.String()
		} else {
			config.Clients[c].ConsentMode = oidc.ClientConsentModeExplicit.String()
		}
	case utils.IsStringInSlice(config.Clients[c].ConsentMode, validOIDCClientConsentModes):
		break
	default:
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidConsentMode, config.Clients[c].ID, strings.Join(append(validOIDCClientConsentModes, "auto"), "', '"), config.Clients[c].ConsentMode))
	}

	if config.Clients[c].ConsentMode == oidc.ClientConsentModePreConfigured.String() && config.Clients[c].ConsentPreConfiguredDuration == nil {
		config.Clients[c].ConsentPreConfiguredDuration = schema.DefaultOpenIDConnectClientConfiguration.ConsentPreConfiguredDuration
	}
}

func validateOIDCClientScopes(c int, config *schema.OpenIDConnectConfiguration, val *schema.StructValidator) {
	if len(config.Clients[c].Scopes) == 0 {
		config.Clients[c].Scopes = schema.DefaultOpenIDConnectClientConfiguration.Scopes
		return
	}

	if !utils.IsStringInSlice(oidc.ScopeOpenID, config.Clients[c].Scopes) {
		config.Clients[c].Scopes = append(config.Clients[c].Scopes, oidc.ScopeOpenID)
	}

	for _, scope := range config.Clients[c].Scopes {
		if !utils.IsStringInSlice(scope, validOIDCScopes) {
			val.Push(fmt.Errorf(
				errFmtOIDCClientInvalidEntry,
				config.Clients[c].ID, "scopes", strings.Join(validOIDCScopes, "', '"), scope))
		}
	}
}

func validateOIDCClientGrantTypes(c int, config *schema.OpenIDConnectConfiguration, val *schema.StructValidator) {
	if len(config.Clients[c].GrantTypes) == 0 {
		config.Clients[c].GrantTypes = schema.DefaultOpenIDConnectClientConfiguration.GrantTypes
		return
	}

	for _, grantType := range config.Clients[c].GrantTypes {
		if !utils.IsStringInSlice(grantType, validOIDCGrantTypes) {
			val.Push(fmt.Errorf(
				errFmtOIDCClientInvalidEntry,
				config.Clients[c].ID, "grant_types", strings.Join(validOIDCGrantTypes, "', '"), grantType))
		}
	}
}

func validateOIDCClientResponseTypes(c int, config *schema.OpenIDConnectConfiguration, _ *schema.StructValidator) {
	if len(config.Clients[c].ResponseTypes) == 0 {
		config.Clients[c].ResponseTypes = schema.DefaultOpenIDConnectClientConfiguration.ResponseTypes
		return
	}
}

func validateOIDCClientResponseModes(c int, config *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	if len(config.Clients[c].ResponseModes) == 0 {
		config.Clients[c].ResponseModes = schema.DefaultOpenIDConnectClientConfiguration.ResponseModes
		return
	}

	for _, responseMode := range config.Clients[c].ResponseModes {
		if !utils.IsStringInSlice(responseMode, validOIDCResponseModes) {
			validator.Push(fmt.Errorf(
				errFmtOIDCClientInvalidEntry,
				config.Clients[c].ID, "response_modes", strings.Join(validOIDCResponseModes, "', '"), responseMode))
		}
	}
}

func validateOIDDClientUserinfoAlgorithm(c int, config *schema.OpenIDConnectConfiguration, val *schema.StructValidator) {
	if config.Clients[c].UserinfoSigningAlgorithm == "" {
		config.Clients[c].UserinfoSigningAlgorithm = schema.DefaultOpenIDConnectClientConfiguration.UserinfoSigningAlgorithm
	} else if !utils.IsStringInSlice(config.Clients[c].UserinfoSigningAlgorithm, validOIDCUserinfoAlgorithms) {
		val.Push(fmt.Errorf(errFmtOIDCClientInvalidUserinfoAlgorithm,
			config.Clients[c].ID, strings.Join(validOIDCUserinfoAlgorithms, ", "), config.Clients[c].UserinfoSigningAlgorithm))
	}
}

func validateOIDCClientRedirectURIs(client schema.OpenIDConnectClientConfiguration, val *schema.StructValidator) {
	for _, redirectURI := range client.RedirectURIs {
		if redirectURI == oauth2InstalledApp {
			if client.Public {
				continue
			}

			val.Push(fmt.Errorf(errFmtOIDCClientRedirectURIPublic, client.ID, oauth2InstalledApp))

			continue
		}

		parsedURL, err := url.Parse(redirectURI)
		if err != nil {
			val.Push(fmt.Errorf(errFmtOIDCClientRedirectURICantBeParsed, client.ID, redirectURI, err))
			continue
		}

		if !parsedURL.IsAbs() || (!client.Public && parsedURL.Scheme == "") {
			val.Push(fmt.Errorf(errFmtOIDCClientRedirectURIAbsolute, client.ID, redirectURI))
			return
		}
	}
}
