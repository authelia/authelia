package validator

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateIdentityProviders validates and updates the IdentityProviders configuration.
func ValidateIdentityProviders(ctx *ValidateCtx, config *schema.Configuration, validator *schema.StructValidator) {
	validateOIDC(ctx, config, validator)
}

func validateOIDC(ctx *ValidateCtx, config *schema.Configuration, validator *schema.StructValidator) {
	if config == nil || config.IdentityProviders.OIDC == nil {
		return
	}

	config.IdentityProviders.OIDC.Discovery.Scopes = append(config.IdentityProviders.OIDC.Discovery.Scopes, validOIDCClientScopes...)

	setOIDCDefaults(config)

	validateOIDCIssuer(config.IdentityProviders.OIDC, validator)
	validateOIDCAuthorizationPolicies(config, validator)
	validateOIDCLifespans(config, validator)
	validateOIDCClaims(config, validator)
	validateOIDCScopes(config, validator)

	sort.Sort(oidc.SortedSigningAlgs(config.IdentityProviders.OIDC.Discovery.ResponseObjectSigningAlgs))

	switch {
	case config.IdentityProviders.OIDC.MinimumParameterEntropy == -1:
		validator.PushWarning(errors.New(errFmtOIDCProviderInsecureDisabledParameterEntropy))
	case config.IdentityProviders.OIDC.MinimumParameterEntropy <= 0:
		config.IdentityProviders.OIDC.MinimumParameterEntropy = oauthelia2.MinParameterEntropy
	case config.IdentityProviders.OIDC.MinimumParameterEntropy < oauthelia2.MinParameterEntropy:
		validator.PushWarning(fmt.Errorf(errFmtOIDCProviderInsecureParameterEntropyUnsafe, oauthelia2.MinParameterEntropy, config.IdentityProviders.OIDC.MinimumParameterEntropy))
	}

	switch config.IdentityProviders.OIDC.EnforcePKCE {
	case "always", "never", "public_clients_only":
		break
	default:
		validator.Push(fmt.Errorf(errFmtOIDCProviderEnforcePKCEInvalidValue, config.IdentityProviders.OIDC.EnforcePKCE))
	}

	validateOIDCOptionsCORS(config.IdentityProviders.OIDC, validator)

	if len(config.IdentityProviders.OIDC.Clients) == 0 {
		validator.Push(errors.New(errFmtOIDCProviderNoClientsConfigured))
	} else {
		validateOIDCClients(ctx, config.IdentityProviders.OIDC, validator)
	}
}

func validateOIDCAuthorizationPolicies(config *schema.Configuration, validator *schema.StructValidator) {
	config.IdentityProviders.OIDC.Discovery.AuthorizationPolicies = []string{policyOneFactor, policyTwoFactor}

	for name, policy := range config.IdentityProviders.OIDC.AuthorizationPolicies {
		add := true

		switch name {
		case "":
			validator.Push(errors.New(errFmtOIDCPolicyInvalidName))

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

		for i := range policy.Rules {
			validateOIDCAuthorizationPoliciesRule(name, i, config, validator)
		}

		config.IdentityProviders.OIDC.AuthorizationPolicies[name] = policy

		if add {
			config.IdentityProviders.OIDC.Discovery.AuthorizationPolicies = append(config.IdentityProviders.OIDC.Discovery.AuthorizationPolicies, name)
		}
	}
}

func validateOIDCAuthorizationPoliciesRule(name string, i int, config *schema.Configuration, validator *schema.StructValidator) {
	switch config.IdentityProviders.OIDC.AuthorizationPolicies[name].Rules[i].Policy {
	case "":
		config.IdentityProviders.OIDC.AuthorizationPolicies[name].Rules[i].Policy = schema.DefaultOpenIDConnectPolicyConfiguration.DefaultPolicy
	case policyOneFactor, policyTwoFactor, policyDeny:
		break
	default:
		validator.Push(fmt.Errorf(errFmtOIDCPolicyRuleInvalidPolicy, name, i+1, utils.StringJoinAnd([]string{policyOneFactor, policyTwoFactor, policyDeny}), config.IdentityProviders.OIDC.AuthorizationPolicies[name].Rules[i].Policy))
	}

	if len(config.IdentityProviders.OIDC.AuthorizationPolicies[name].Rules[i].Subjects) == 0 && len(config.IdentityProviders.OIDC.AuthorizationPolicies[name].Rules[i].Networks) == 0 {
		validator.Push(fmt.Errorf(errFmtOIDCPolicyRuleMissingOption, name, i+1))

		return
	}

	for _, subjectRule := range config.IdentityProviders.OIDC.AuthorizationPolicies[name].Rules[i].Subjects {
		for _, subject := range subjectRule {
			if !IsSubjectValidBasic(subject) {
				validator.Push(fmt.Errorf(errFmtOIDCPolicyRuleInvalidSubject, name, i+1, subject))
			}
		}
	}
}

func validateOIDCLifespans(config *schema.Configuration, _ *schema.StructValidator) {
	for name := range config.IdentityProviders.OIDC.Lifespans.Custom {
		config.IdentityProviders.OIDC.Discovery.Lifespans = append(config.IdentityProviders.OIDC.Discovery.Lifespans, name)
	}
}

//nolint:gocyclo
func validateOIDCClaims(config *schema.Configuration, validator *schema.StructValidator) {
	for name, policy := range config.IdentityProviders.OIDC.ClaimsPolicies {
		var claims []string

		seen := make(map[string]string, len(policy.CustomClaims))

		for key, properties := range policy.CustomClaims {
			if properties.Name == "" {
				properties.Name = key
				policy.CustomClaims[key] = properties
			}

			if properties.Attribute == "" {
				properties.Attribute = key
				policy.CustomClaims[key] = properties
			}

			if utils.IsStringInSlice(properties.Name, validOIDCReservedClaims) {
				validator.Push(fmt.Errorf("identity_providers: oidc: claims_policies: %s: custom_claims: claim with name '%s' can't be used in a claims policy as it's a standard claim", name, properties.Name))
			}

			if !utils.IsStringInSlice(properties.Name, claims) {
				claims = append(claims, properties.Name)
			}

			if !utils.IsStringInSlice(properties.Name, config.IdentityProviders.OIDC.Discovery.Claims) {
				config.IdentityProviders.OIDC.Discovery.Claims = append(config.IdentityProviders.OIDC.Discovery.Claims, properties.Name)
			}

			if k, ok := seen[properties.Name]; ok {
				validator.Push(fmt.Errorf("identity_providers: oidc: claims_policies: %s: custom_claims: claim with name '%s' is mapped in both '%s' and the '%s' claim configurations", name, properties.Name, key, k))
			} else {
				seen[properties.Name] = key
			}

			if !isUserAttributeValid(properties.Attribute, config) {
				validator.Push(fmt.Errorf("identity_providers: oidc: claims_policies: %s: claim with name '%s' has an attribute name '%s' which is not a known attribute", name, properties.Name, properties.Attribute))
			}
		}

		for _, claim := range policy.IDToken {
			if utils.IsStringInSlice(claim, validOIDCReservedIDTokenClaims) {
				validator.Push(fmt.Errorf("identity_providers: oidc: claims_policies: %s: id_token: claim with name '%s' can't be used in a claims policy as it's a standard claim", name, claim))
			} else if !utils.IsStringInSlice(claim, claims) && !utils.IsStringInSlice(claim, validOIDCClientClaims) {
				validator.Push(fmt.Errorf("identity_providers: oidc: claims_policies: %s: id_token: claim with name '%s' is not known", name, claim))
			}
		}

		switch policy.IDTokenAudienceMode {
		case "":
			policy.IDTokenAudienceMode = oidc.IDTokenAudienceModeSpecification
		case oidc.IDTokenAudienceModeSpecification, oidc.IDTokenAudienceModeExperimentalMerged:
			break
		default:
			validator.Push(fmt.Errorf("identity_providers: oidc: claims_policies: %s: option 'id_token_audience_mode' must be one of %s but it's configured as '%s'", name, utils.StringJoinOr([]string{oidc.IDTokenAudienceModeSpecification, oidc.IDTokenAudienceModeExperimentalMerged}), policy.IDTokenAudienceMode))
		}

		for _, claim := range policy.AccessToken {
			if utils.IsStringInSlice(claim, validOIDCReservedClaims) {
				validator.Push(fmt.Errorf("identity_providers: oidc: claims_policies: %s: access_token: claim with name '%s' can't be used in a claims policy as it's a standard claim", name, claim))
			} else if !utils.IsStringInSlice(claim, claims) && !utils.IsStringInSlice(claim, validOIDCClientClaims) {
				validator.Push(fmt.Errorf("identity_providers: oidc: claims_policies: %s: access_token: claim with name '%s' is not known", name, claim))
			}
		}

		config.IdentityProviders.OIDC.ClaimsPolicies[name] = policy
	}
}

func validateOIDCScopes(config *schema.Configuration, validator *schema.StructValidator) {
	for scope, properties := range config.IdentityProviders.OIDC.Scopes {
		if utils.IsStringInSlice(scope, validOIDCClientScopes) {
			validator.Push(fmt.Errorf("identity_providers: oidc: scopes: scope with name '%s' can't be used as a custom scope because it's a standard scope", scope))
		} else if strings.HasPrefix(scope, "authelia.") {
			validator.Push(fmt.Errorf("identity_providers: oidc: scopes: scope with name '%s' can't be used as a custom scope because all scopes prefixed with 'authelia.' are reserved", scope))
		}

		if !utils.IsStringInSlice(scope, config.IdentityProviders.OIDC.Discovery.Scopes) {
			config.IdentityProviders.OIDC.Discovery.Scopes = append(config.IdentityProviders.OIDC.Discovery.Scopes, scope)
		}

		for _, claim := range properties.Claims {
			if utils.IsStringInSlice(claim, validOIDCReservedClaims) {
				validator.Push(fmt.Errorf("identity_providers: oidc: scopes: %s: claim with name '%s' can't be used in a custom scope as it's a standard claim", scope, claim))
			} else if !utils.IsStringInSlice(claim, config.IdentityProviders.OIDC.Discovery.Claims) && !utils.IsStringInSlice(claim, validOIDCClientClaims) {
				validator.Push(fmt.Errorf("identity_providers: oidc: scopes: %s: claim with name '%s' is not a known claim", scope, claim))
			}
		}
	}
}

//nolint:gocyclo
func isUserAttributeValid(name string, config *schema.Configuration) (valid bool) {
	if _, ok := config.Definitions.UserAttributes[name]; ok {
		return true
	}

	if config.AuthenticationBackend.LDAP != nil {
		switch name {
		case attributeUserUsername, attributeUserDisplayName, attributeUserGroups, attributeUserEmail, attributeUserEmails:
			return true
		case attributeUserGivenName:
			return config.AuthenticationBackend.LDAP.Attributes.GivenName != ""
		case attributeUserMiddleName:
			return config.AuthenticationBackend.LDAP.Attributes.MiddleName != ""
		case attributeUserFamilyName:
			return config.AuthenticationBackend.LDAP.Attributes.FamilyName != ""
		case attributeUserNickname:
			return config.AuthenticationBackend.LDAP.Attributes.Nickname != ""
		case attributeUserProfile:
			return config.AuthenticationBackend.LDAP.Attributes.Profile != ""
		case attributeUserPicture:
			return config.AuthenticationBackend.LDAP.Attributes.Picture != ""
		case attributeUserWebsite:
			return config.AuthenticationBackend.LDAP.Attributes.Website != ""
		case attributeUserGender:
			return config.AuthenticationBackend.LDAP.Attributes.Gender != ""
		case attributeUserBirthdate:
			return config.AuthenticationBackend.LDAP.Attributes.Birthdate != ""
		case attributeUserZoneInfo:
			return config.AuthenticationBackend.LDAP.Attributes.ZoneInfo != ""
		case attributeUserLocale:
			return config.AuthenticationBackend.LDAP.Attributes.Locale != ""
		case attributeUserPhoneNumber:
			return config.AuthenticationBackend.LDAP.Attributes.PhoneNumber != ""
		case attributeUserPhoneExtension:
			return config.AuthenticationBackend.LDAP.Attributes.PhoneExtension != ""
		case attributeUserStreetAddress:
			return config.AuthenticationBackend.LDAP.Attributes.StreetAddress != ""
		case attributeUserLocality:
			return config.AuthenticationBackend.LDAP.Attributes.Locality != ""
		case attributeUserRegion:
			return config.AuthenticationBackend.LDAP.Attributes.Region != ""
		case attributeUserPostalCode:
			return config.AuthenticationBackend.LDAP.Attributes.PostalCode != ""
		case attributeUserCountry:
			return config.AuthenticationBackend.LDAP.Attributes.Country != ""
		default:
			if config.AuthenticationBackend.LDAP.Attributes.Extra == nil {
				return false
			}

			for key, attr := range config.AuthenticationBackend.LDAP.Attributes.Extra {
				if attr.Name != "" {
					if attr.Name == name {
						return true
					}
				} else if key == name {
					return true
				}
			}

			return false
		}
	}

	if utils.IsStringInSlice(name, validUserAttributes) {
		return true
	}

	if config.AuthenticationBackend.File == nil {
		return false
	}

	if _, ok := config.AuthenticationBackend.File.ExtraAttributes[name]; ok {
		return true
	}

	return false
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
		validator.Push(errors.New(errFmtOIDCProviderNoPrivateKey))
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

//nolint:gocyclo
func validateOIDCIssuerJSONWebKeys(config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	var (
		props *JWKProperties
		err   error
	)

	config.Discovery.ResponseObjectSigningKeyIDs = []string{}
	config.Discovery.ResponseObjectEncryptionKeyIDs = []string{}
	config.Discovery.DefaultSigKeyIDs = map[string]string{}
	config.Discovery.DefaultEncKeyIDs = map[string]string{}

	for i := 0; i < len(config.JSONWebKeys); i++ {
		if config.JSONWebKeys[i].Key == nil {
			if len(config.JSONWebKeys[i].KeyID) != 0 {
				validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysWithKeyID, i+1, config.JSONWebKeys[i].KeyID))
			} else {
				validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysMissing, i+1))
			}

			continue
		}

		if key, ok := config.JSONWebKeys[i].Key.(*rsa.PrivateKey); ok && key.N == nil {
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

		switch config.JSONWebKeys[i].Use {
		case oidc.KeyUseEncryption:
			if config.JSONWebKeys[i].KeyID != "" && utils.IsStringInSlice(config.JSONWebKeys[i].KeyID, config.Discovery.ResponseObjectEncryptionKeyIDs) {
				validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysAttributeNotUnique, i+1, config.JSONWebKeys[i].KeyID, attrOIDCKeyID))
			}

			config.Discovery.ResponseObjectEncryptionKeyIDs = append(config.Discovery.ResponseObjectEncryptionKeyIDs, config.JSONWebKeys[i].KeyID)
		default:
			if config.JSONWebKeys[i].KeyID != "" && utils.IsStringInSlice(config.JSONWebKeys[i].KeyID, config.Discovery.ResponseObjectSigningKeyIDs) {
				validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysAttributeNotUnique, i+1, config.JSONWebKeys[i].KeyID, attrOIDCKeyID))
			}

			config.Discovery.ResponseObjectSigningKeyIDs = append(config.Discovery.ResponseObjectSigningKeyIDs, config.JSONWebKeys[i].KeyID)
		}

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
	if config.JSONWebKeys[i].Use == "" {
		config.JSONWebKeys[i].Use = props.Use
	}

	switch config.JSONWebKeys[i].Use {
	case oidc.KeyUseSignature:
		validateOIDCIssuerPrivateKeysSigAlg(i, props, config, validator)
	case oidc.KeyUseEncryption:
		validateOIDCIssuerPrivateKeysEncAlg(i, props, config, validator)
	default:
		validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysInvalidOptionOneOf, i+1, config.JSONWebKeys[i].KeyID, attrOIDCKeyUse, utils.StringJoinOr([]string{oidc.KeyUseSignature, oidc.KeyUseEncryption}), config.JSONWebKeys[i].Use))
	}
}

func validateOIDCIssuerPrivateKeysSigAlg(i int, props *JWKProperties, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	switch {
	case config.JSONWebKeys[i].Algorithm == "":
		config.JSONWebKeys[i].Algorithm = props.Algorithm

		fallthrough
	case utils.IsStringInSlice(config.JSONWebKeys[i].Algorithm, validOIDCIssuerJWKSigningAlgs):
		if config.JSONWebKeys[i].KeyID != "" && config.JSONWebKeys[i].Algorithm != "" {
			if _, ok := config.Discovery.DefaultSigKeyIDs[config.JSONWebKeys[i].Algorithm]; !ok {
				config.Discovery.DefaultSigKeyIDs[config.JSONWebKeys[i].Algorithm] = config.JSONWebKeys[i].KeyID
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

func validateOIDCIssuerPrivateKeysEncAlg(i int, props *JWKProperties, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	switch {
	case config.JSONWebKeys[i].Algorithm == "":
		config.JSONWebKeys[i].Algorithm = props.Algorithm

		fallthrough
	case utils.IsStringInSlice(config.JSONWebKeys[i].Algorithm, validOIDCJWKEncryptionAlgs):
		if config.JSONWebKeys[i].KeyID != "" && config.JSONWebKeys[i].Algorithm != "" {
			if _, ok := config.Discovery.DefaultEncKeyIDs[config.JSONWebKeys[i].Algorithm]; !ok {
				config.Discovery.DefaultEncKeyIDs[config.JSONWebKeys[i].Algorithm] = config.JSONWebKeys[i].KeyID
			}
		}
	default:
		validator.Push(fmt.Errorf(errFmtOIDCProviderPrivateKeysInvalidOptionOneOf, i+1, config.JSONWebKeys[i].KeyID, attrOIDCAlgorithm, utils.StringJoinOr(validOIDCJWKEncryptionAlgs), config.JSONWebKeys[i].Algorithm))
	}

	if config.JSONWebKeys[i].Algorithm != "" {
		if !utils.IsStringInSlice(config.JSONWebKeys[i].Algorithm, config.Discovery.ResponseObjectEncryptionAlgs) {
			config.Discovery.ResponseObjectEncryptionAlgs = append(config.Discovery.ResponseObjectEncryptionAlgs, config.JSONWebKeys[i].Algorithm)
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

func setOIDCDefaults(config *schema.Configuration) {
	if config.IdentityProviders.OIDC.Lifespans.AccessToken == durationZero {
		config.IdentityProviders.OIDC.Lifespans.AccessToken = schema.DefaultOpenIDConnectConfiguration.Lifespans.AccessToken
	}

	if config.IdentityProviders.OIDC.Lifespans.RefreshToken == durationZero {
		config.IdentityProviders.OIDC.Lifespans.RefreshToken = schema.DefaultOpenIDConnectConfiguration.Lifespans.RefreshToken
	}

	if config.IdentityProviders.OIDC.Lifespans.IDToken == durationZero {
		config.IdentityProviders.OIDC.Lifespans.IDToken = schema.DefaultOpenIDConnectConfiguration.Lifespans.IDToken
	}

	if config.IdentityProviders.OIDC.Lifespans.AuthorizeCode == durationZero {
		config.IdentityProviders.OIDC.Lifespans.AuthorizeCode = schema.DefaultOpenIDConnectConfiguration.Lifespans.AuthorizeCode
	}

	if config.IdentityProviders.OIDC.Lifespans.DeviceCode == durationZero {
		config.IdentityProviders.OIDC.Lifespans.DeviceCode = schema.DefaultOpenIDConnectConfiguration.Lifespans.DeviceCode
	}

	if config.IdentityProviders.OIDC.EnforcePKCE == "" {
		config.IdentityProviders.OIDC.EnforcePKCE = schema.DefaultOpenIDConnectConfiguration.EnforcePKCE
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
				validator.Push(errors.New(errFmtOIDCCORSInvalidOriginWildcard))
			}

			if config.CORS.AllowedOriginsFromClientRedirectURIs {
				validator.Push(errors.New(errFmtOIDCCORSInvalidOriginWildcardWithClients))
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
		validator.PushWarning(errors.New(errFmtOIDCClientsDeprecated))
	}

	if len(blankClientIDs) != 0 {
		validator.Push(fmt.Errorf(errFmtOIDCClientsWithEmptyID, utils.StringJoinBuild(", ", "or", "", blankClientIDs)))
	}

	if len(duplicateClientIDs) != 0 {
		validator.Push(fmt.Errorf(errFmtOIDCClientsDuplicateID, utils.StringJoinOr(duplicateClientIDs)))
	}

	ctx.cacheSectorIdentifierURIs = nil
}

//nolint:gocyclo
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
	validateOIDDClientEncryptionAlgs(c, config, validator)

	validateOIDCClientSectorIdentifier(ctx, c, config, validator, errDeprecatedFunc)

	validateOIDCClientPublicKeys(c, config, validator)

	var (
		method, alg                                  string
		confidential, public, econfidential, epublic bool
	)

	method, alg, econfidential, epublic = validateOIDCClientEndpointAuth(c, config, attrOIDCTokenAuthMethod, config.Clients[c].TokenEndpointAuthMethod, attrOIDCTokenAuthSigningAlg, config.Clients[c].TokenEndpointAuthSigningAlg, validator)
	confidential = boolApply(confidential, econfidential)
	public = boolApply(public, epublic)

	config.Clients[c].TokenEndpointAuthMethod = method
	config.Clients[c].TokenEndpointAuthSigningAlg = alg

	method, alg, econfidential, epublic = validateOIDCClientEndpointAuth(c, config, attrOIDCRevocationAuthMethod, config.Clients[c].RevocationEndpointAuthMethod, attrOIDCRevocationAuthSigningAlg, config.Clients[c].RevocationEndpointAuthSigningAlg, validator)
	confidential = boolApply(confidential, econfidential)
	public = boolApply(public, epublic)

	config.Clients[c].RevocationEndpointAuthMethod = method
	config.Clients[c].RevocationEndpointAuthSigningAlg = alg

	method, alg, econfidential, epublic = validateOIDCClientEndpointAuth(c, config, attrOIDCIntrospectionAuthMethod, config.Clients[c].IntrospectionEndpointAuthMethod, attrOIDCIntrospectionAuthSigningAlg, config.Clients[c].IntrospectionEndpointAuthSigningAlg, validator)
	confidential = boolApply(confidential, econfidential)
	public = boolApply(public, epublic)

	config.Clients[c].IntrospectionEndpointAuthMethod = method
	config.Clients[c].IntrospectionEndpointAuthSigningAlg = alg

	method, alg, econfidential, epublic = validateOIDCClientEndpointAuth(c, config, attrOIDCPARAuthMethod, config.Clients[c].PushedAuthorizationRequestEndpointAuthMethod, attrOIDCPARAuthSigningAlg, config.Clients[c].PushedAuthorizationRequestAuthSigningAlg, validator)
	confidential = boolApply(confidential, econfidential)
	public = boolApply(public, epublic)

	if config.Clients[c].Discovery.ClientSecretPlainText && !config.Clients[c].Discovery.RequestObjectSymmetricSigEncAlg && !config.Clients[c].Discovery.ResponseObjectSymmetricSigEncAlg {
		validator.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidSecretPlainText, config.Clients[c].ID))
	}

	config.Clients[c].PushedAuthorizationRequestEndpointAuthMethod = method
	config.Clients[c].PushedAuthorizationRequestAuthSigningAlg = alg

	if confidential {
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidSecret, config.Clients[c].ID))
	}

	if public {
		validator.Push(fmt.Errorf(errFmtOIDCClientPublicInvalidSecret, config.Clients[c].ID))
	}
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

		if config.Clients[c].JSONWebKeys[i].Key == nil {
			if len(config.Clients[c].JSONWebKeys[i].KeyID) != 0 {
				validator.Push(fmt.Errorf(errFmtOIDCClientPublicKeysWithIDInvalidOptionMissingOneOf, config.Clients[c].ID, i+1, config.Clients[c].JSONWebKeys[i].KeyID, attrOIDCKey))
			} else {
				validator.Push(fmt.Errorf(errFmtOIDCClientPublicKeysInvalidOptionMissingOneOf, config.Clients[c].ID, i+1, attrOIDCKey))
			}

			continue
		}

		validateOIDCClientJSONWebKeysListKeyUseAlg(c, i, props, config, validator)

		var checkEqualKey bool

		switch key := config.Clients[c].JSONWebKeys[i].Key.(type) {
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
	case oidc.KeyUseSignature, oidc.KeyUseEncryption:
		break
	default:
		validator.Push(fmt.Errorf(errFmtOIDCClientPublicKeysInvalidOptionOneOf, config.Clients[c].ID, i+1, config.Clients[c].JSONWebKeys[i].KeyID, attrOIDCKeyUse, utils.StringJoinOr([]string{oidc.KeyUseSignature, oidc.KeyUseEncryption}), config.Clients[c].JSONWebKeys[i].Use))
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

//nolint:gocyclo
func validateOIDCClientScopes(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator, ccg bool, errDeprecatedFunc func()) {
	if len(config.Clients[c].Scopes) == 0 && !ccg {
		config.Clients[c].Scopes = schema.DefaultOpenIDConnectClientConfiguration.Scopes
	}

	invalid, duplicates := validateList(config.Clients[c].Scopes, config.Discovery.Scopes, true)

	if len(duplicates) != 0 {
		errDeprecatedFunc()

		validator.PushWarning(fmt.Errorf(errFmtOIDCClientInvalidEntryDuplicates, config.Clients[c].ID, attrOIDCScopes, utils.StringJoinAnd(duplicates)))
	}

	if ccg {
		validateOIDCClientScopesClientCredentialsGrant(c, config, validator)
	} else if len(invalid) != 0 {
		validator.PushWarning(fmt.Errorf(errFmtOIDCClientUnknownScopeEntries, config.Clients[c].ID, attrOIDCScopes, utils.StringJoinOr(config.Discovery.Scopes), utils.StringJoinAnd(invalid)))
	}

	var claims *schema.IdentityProvidersOpenIDConnectClaimsPolicy

	if config.Clients[c].ClaimsPolicy != "" {
		if v, ok := config.ClaimsPolicies[config.Clients[c].ClaimsPolicy]; ok {
			claims = &v
		} else {
			validator.Push(fmt.Errorf("identity_providers: oidc: clients: client '%s': option '%s' is not the name of a configured claims policy", config.Clients[c].ID, "claims_policy"))
		}
	}

	for _, scope := range config.Clients[c].Scopes {
		if utils.IsStringInSlice(scope, validOIDCClientScopes) {
			continue
		}

		properties, ok := config.Scopes[scope]
		if !ok {
			continue
		}

		if claims == nil {
			continue
		}

		for _, claim := range properties.Claims {
			if utils.IsStringInSlice(claim, validOIDCClientClaims) {
				continue
			}

			if mapping := claims.CustomClaims.GetCustomClaimByName(claim); mapping.Name == "" {
				validator.Push(fmt.Errorf("identity_providers: oidc: clients: client '%s': option 'scopes' contains value '%s' which requires claim '%s' but the claim is not a claim provided by 'claims_policy' with the name '%s' or a standard claim", config.Clients[c].ID, scope, claim, config.Clients[c].ClaimsPolicy))
			}
		}
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
func validateOIDCClientEndpointAuth(c int, config *schema.IdentityProvidersOpenIDConnect, keyMethod, valueMethod, keyAlg, valueAlg string, validator *schema.StructValidator) (method, alg string, secretConfidential, secretPublic bool) {
	implicit := len(config.Clients[c].ResponseTypes) != 0 && utils.IsStringSliceContainsAll(config.Clients[c].ResponseTypes, validOIDCClientResponseTypesImplicitFlow)

	if config.Clients[c].Secret.Valid() {
		config.Clients[c].Discovery.ClientSecretPlainText = config.Clients[c].Secret.IsPlainText()
	}

	switch {
	case valueMethod == "":
		if config.Clients[c].Public {
			valueMethod = oidc.ClientAuthMethodNone
		} else {
			valueMethod = oidc.ClientAuthMethodClientSecretBasic
		}
	case !utils.IsStringInSlice(valueMethod, validOIDCClientTokenEndpointAuthMethods):
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
			config.Clients[c].ID, keyMethod, utils.StringJoinOr(validOIDCClientTokenEndpointAuthMethods), valueMethod))

		return valueMethod, valueAlg, secretConfidential, secretPublic
	case valueMethod == oidc.ClientAuthMethodNone && !config.Clients[c].Public && !implicit:
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidEndpointAuthMethod,
			config.Clients[c].ID, keyMethod, utils.StringJoinOr(validOIDCClientTokenEndpointAuthMethodsConfidential), utils.StringJoinAnd(validOIDCClientResponseTypesImplicitFlow), valueMethod))
	case valueMethod != oidc.ClientAuthMethodNone && config.Clients[c].Public:
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidEndpointAuthMethodPublic,
			config.Clients[c].ID, keyMethod, valueMethod))
	}

	secret := false

	switch valueMethod {
	case oidc.ClientAuthMethodClientSecretJWT:
		valueAlg = validateOIDCClientEndpointAuthClientSecretJWT(c, config, keyMethod, valueMethod, keyAlg, valueAlg, validator)

		secret = true
	case oidc.ClientAuthMethodClientSecretPost, oidc.ClientAuthMethodClientSecretBasic:
		secret = true
	case oidc.ClientAuthMethodPrivateKeyJWT:
		validateOIDCClientEndpointAuthPublicKeyJWT(c, config, keyMethod, valueMethod, keyAlg, valueAlg, validator)
	}

	if secret {
		if config.Clients[c].Public {
			return valueMethod, valueAlg, secretConfidential, secretPublic
		}

		if !config.Clients[c].Secret.Valid() {
			secretConfidential = true
		} else if valueMethod == oidc.ClientAuthMethodClientSecretJWT {
			config.Clients[c].Discovery.RequestObjectSymmetricSigEncAlg = true

			if !config.Clients[c].Discovery.ClientSecretPlainText {
				validator.Push(fmt.Errorf(errFmtOIDCClientInvalidSecretNotPlainText, config.Clients[c].ID, keyMethod, oidc.ClientAuthMethodClientSecretJWT))
			}
		}
	} else if config.Clients[c].Secret != nil {
		if config.Clients[c].Public {
			secretPublic = true
		} else {
			validator.Push(fmt.Errorf(errFmtOIDCClientPublicInvalidSecretClientAuthMethod, config.Clients[c].ID, keyMethod, valueMethod))
		}
	}

	return valueMethod, valueAlg, secretConfidential, secretPublic
}

func validateOIDCClientEndpointAuthClientSecretJWT(c int, config *schema.IdentityProvidersOpenIDConnect, keyMethod, valueMethod, keyAlg, valueAlg string, validator *schema.StructValidator) (alg string) {
	switch {
	case valueAlg == "":
		valueAlg = oidc.SigningAlgHMACUsingSHA256
	case !utils.IsStringInSlice(valueAlg, validOIDCClientTokenEndpointAuthSigAlgsClientSecretJWT):
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidEndpointAuthSigAlg, config.Clients[c].ID, keyAlg, utils.StringJoinOr(validOIDCClientTokenEndpointAuthSigAlgsClientSecretJWT), keyMethod, valueMethod))
	}

	return valueAlg
}

func validateOIDCClientEndpointAuthPublicKeyJWT(c int, config *schema.IdentityProvidersOpenIDConnect, keyMethod, valueMethod, keyAlg, valueAlg string, validator *schema.StructValidator) {
	switch {
	case valueAlg == "":
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidTokenEndpointAuthSigAlgMissingPrivateKeyJWT, config.Clients[c].ID))
	case !utils.IsStringInSlice(valueAlg, validOIDCIssuerJWKSigningAlgs):
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidEndpointAuthSigAlg, config.Clients[c].ID, keyAlg, utils.StringJoinOr(validOIDCIssuerJWKSigningAlgs), keyMethod, valueMethod))
	}

	if config.Clients[c].JSONWebKeysURI == nil {
		if len(config.Clients[c].JSONWebKeys) == 0 {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidPublicKeysPrivateKeyJWT, config.Clients[c].ID))
		} else if len(config.Clients[c].Discovery.RequestObjectSigningAlgs) != 0 && !utils.IsStringInSlice(valueAlg, config.Clients[c].Discovery.RequestObjectSigningAlgs) {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidTokenEndpointAuthSigAlgReg, config.Clients[c].ID, utils.StringJoinOr(config.Clients[c].Discovery.RequestObjectSigningAlgs), valueMethod))
		}
	}
}

func validateOIDDClientSigningAlgs(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	if config.Clients[c].Secret.Valid() {
		config.Clients[c].Discovery.ClientSecretPlainText = config.Clients[c].Secret.IsPlainText()
	}

	config.Clients[c].AuthorizationSignedResponseAlg, config.Clients[c].AuthorizationSignedResponseKeyID = validateOIDDClientSigningAlg(c, config, config.Clients[c].AuthorizationSignedResponseAlg, schema.DefaultOpenIDConnectClientConfiguration.AuthorizationSignedResponseAlg, config.Clients[c].AuthorizationSignedResponseKeyID, attrOIDCAuthorizationPrefix, validator)
	config.Clients[c].IDTokenSignedResponseAlg, config.Clients[c].IDTokenSignedResponseKeyID = validateOIDDClientSigningAlg(c, config, config.Clients[c].IDTokenSignedResponseAlg, schema.DefaultOpenIDConnectClientConfiguration.IDTokenSignedResponseAlg, config.Clients[c].IDTokenSignedResponseKeyID, attrOIDCIDTokenPrefix, validator)
	config.Clients[c].AccessTokenSignedResponseAlg, config.Clients[c].AccessTokenSignedResponseKeyID = validateOIDDClientSigningAlg(c, config, config.Clients[c].AccessTokenSignedResponseAlg, schema.DefaultOpenIDConnectClientConfiguration.AccessTokenSignedResponseAlg, config.Clients[c].AccessTokenSignedResponseKeyID, attrOIDCAccessTokenPrefix, validator)
	config.Clients[c].UserinfoSignedResponseAlg, config.Clients[c].UserinfoSignedResponseKeyID = validateOIDDClientSigningAlg(c, config, config.Clients[c].UserinfoSignedResponseAlg, schema.DefaultOpenIDConnectClientConfiguration.UserinfoSignedResponseAlg, config.Clients[c].UserinfoSignedResponseKeyID, attrOIDCUserinfoPrefix, validator)
	config.Clients[c].IntrospectionSignedResponseAlg, config.Clients[c].IntrospectionSignedResponseKeyID = validateOIDDClientSigningAlg(c, config, config.Clients[c].IntrospectionSignedResponseAlg, schema.DefaultOpenIDConnectClientConfiguration.IntrospectionSignedResponseAlg, config.Clients[c].IntrospectionSignedResponseKeyID, attrOIDCIntrospectionPrefix, validator)
}

func validateOIDDClientSigningAlg(c int, config *schema.IdentityProvidersOpenIDConnect, alg, defaultAlg, kid, attribute string, validator *schema.StructValidator) (outAlg, outKID string) {
	outAlg, outKID = validateOIDCAlgKIDDefault(config, alg, kid, defaultAlg)

	key := fmt.Sprintf("%s_signed_response_alg", attribute)

	if !config.Discovery.JWTResponseAccessTokens && attribute == attrOIDCAccessTokenPrefix {
		switch outAlg {
		case "", oidc.SigningAlgNone:
			break
		default:
			config.Discovery.JWTResponseAccessTokens = true
		}
	}

	switch outKID {
	case "":
		switch outAlg {
		case "", oidc.SigningAlgRSAUsingSHA256:
			break
		case oidc.SigningAlgNone:
			switch attribute {
			case attrOIDCIDTokenPrefix, attrOIDCAuthorizationPrefix:
				validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
					config.Clients[c].ID, key, utils.StringJoinOr(config.Discovery.ResponseObjectSigningAlgs), outAlg))
			}
		default:
			if jwt.IsSignedJWTClientSecretAlgStr(outAlg) {
				config.Clients[c].Discovery.ResponseObjectSymmetricSigEncAlg = true

				if !config.Clients[c].Discovery.ClientSecretPlainText {
					validator.Push(fmt.Errorf(errFmtOIDCClientInvalidSecretNotPlainText, config.Clients[c].ID, key, outAlg))
				}

				break
			}

			if !utils.IsStringInSlice(outAlg, config.Discovery.ResponseObjectSigningAlgs) {
				if attribute == attrOIDCIDTokenPrefix {
					validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
						config.Clients[c].ID, key, utils.StringJoinOr(config.Discovery.ResponseObjectSigningAlgs), outAlg))
				} else {
					validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
						config.Clients[c].ID, key, utils.StringJoinOr(append(config.Discovery.ResponseObjectSigningAlgs, oidc.SigningAlgNone)), outAlg))
				}
			}
		}
	default:
		if !utils.IsStringInSlice(outKID, config.Discovery.ResponseObjectSigningKeyIDs) {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
				config.Clients[c].ID, fmt.Sprintf("%s_signed_response_key_id", attribute), utils.StringJoinOr(config.Discovery.ResponseObjectSigningKeyIDs), outKID))
		} else {
			outAlg = getResponseObjectAlgFromKID(config, outKID, outAlg)
		}
	}

	return
}

func validateOIDDClientEncryptionAlgs(c int, config *schema.IdentityProvidersOpenIDConnect, validator *schema.StructValidator) {
	if config.Clients[c].Secret.Valid() {
		config.Clients[c].Discovery.ClientSecretPlainText = config.Clients[c].Secret.IsPlainText()
	}

	config.Clients[c].AuthorizationEncryptedResponseEnc = validateOIDDClientEncryptionAlg(c, config, config.Clients[c].AuthorizationEncryptedResponseAlg, config.Clients[c].AuthorizationSignedResponseAlg, config.Clients[c].AuthorizationEncryptedResponseEnc, config.Clients[c].AuthorizationEncryptedResponseKeyID, attrOIDCAuthorizationPrefix, validator)
	config.Clients[c].IDTokenEncryptedResponseEnc = validateOIDDClientEncryptionAlg(c, config, config.Clients[c].IDTokenEncryptedResponseAlg, config.Clients[c].IDTokenSignedResponseAlg, config.Clients[c].IDTokenEncryptedResponseEnc, config.Clients[c].IDTokenEncryptedResponseKeyID, attrOIDCIDTokenPrefix, validator)
	config.Clients[c].AccessTokenEncryptedResponseEnc = validateOIDDClientEncryptionAlg(c, config, config.Clients[c].AccessTokenEncryptedResponseAlg, config.Clients[c].AccessTokenSignedResponseAlg, config.Clients[c].AccessTokenEncryptedResponseEnc, config.Clients[c].AccessTokenEncryptedResponseKeyID, attrOIDCAccessTokenPrefix, validator)
	config.Clients[c].UserinfoEncryptedResponseEnc = validateOIDDClientEncryptionAlg(c, config, config.Clients[c].UserinfoEncryptedResponseAlg, config.Clients[c].UserinfoSignedResponseAlg, config.Clients[c].UserinfoEncryptedResponseEnc, config.Clients[c].UserinfoEncryptedResponseKeyID, attrOIDCUserinfoPrefix, validator)
	config.Clients[c].IntrospectionEncryptedResponseEnc = validateOIDDClientEncryptionAlg(c, config, config.Clients[c].IntrospectionEncryptedResponseAlg, config.Clients[c].IntrospectionSignedResponseAlg, config.Clients[c].IntrospectionEncryptedResponseEnc, config.Clients[c].IntrospectionEncryptedResponseKeyID, attrOIDCIntrospectionPrefix, validator)
}

func validateOIDDClientEncryptionAlg(c int, config *schema.IdentityProvidersOpenIDConnect, alg, sigAlg, enc, kid, attribute string, validator *schema.StructValidator) (outEnc string) {
	outEnc = enc

	if len(kid) == 0 && jwt.IsNoneAlg(alg) {
		return
	}

	key := fmt.Sprintf("%s_encrypted_response_alg", attribute)

	if !utils.IsStringInSlice(alg, validOIDCClientJWKEncryptionKeyAlgs) {
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
			config.Clients[c].ID, key, utils.StringJoinOr(validOIDCClientJWKEncryptionKeyAlgs), alg))
	} else {
		if jwt.IsEncryptedJWTClientSecretAlgStr(alg) && !config.Clients[c].Discovery.ClientSecretPlainText {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidSecretNotPlainText, config.Clients[c].ID, key, alg))
		}

		if !jwt.IsEncryptedJWTClientSecretAlgStr(alg) && config.Clients[c].JSONWebKeysURI == nil && len(config.Clients[c].JSONWebKeys) == 0 {
			validator.Push(fmt.Errorf(errFmtOIDCClientOption+"'jwks_uri' or 'jwks' must be configured when '%s_encrypted_response_alg' is set to '%s'", config.Clients[c].ID, attribute, alg))
		}

		if jwt.IsNoneAlg(sigAlg) {
			validator.Push(fmt.Errorf(errFmtOIDCClientOption+"'%s_encrypted_response_alg' must either not be configured or set to 'none' if '%s_signed_response_alg' is set to '%s'", config.Clients[c].ID, attribute, attribute, sigAlg))
		}
	}

	if len(outEnc) == 0 {
		outEnc = oidc.EncryptionEncA128CBCHS256
	}

	if !utils.IsStringInSlice(outEnc, validOIDCClientJWKContentEncryptionAlgs) {
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidValue,
			config.Clients[c].ID, fmt.Sprintf("%s_encrypted_response_enc", attribute), utils.StringJoinOr(validOIDCClientJWKContentEncryptionAlgs), outEnc))
	}

	return
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
