export interface OpenIDConnectClient {
    ID: string;
    Name?: string;
    Public?: boolean; //aka Public

    //this should be ClientSecretDigest but I don't think we need to ever display it beyond creation/regeneration of secrets which can be done separately in plaintext
    //ClientSecret: string ;
    //RotatedClientSecrets?: string[];

    //begin advanced options
    SectorIdentifierURI?: string;
    RequirePushedAuthorizationRequest?: boolean;

    RequirePKCE?: boolean;
    RequirePKCEChallengeMethod?: boolean;
    PKCEChallengeMethod?: string;

    Audience?: string[];
    Scopes: string[];
    RedirectURIs: string[];
    GrantTypes?: string[];
    ResponseTypes?: string[];
    ResponseModes?: ResponseModeType[];

    Lifespan?: IdentityProvidersOpenIDConnectLifespanToken;

    AuthorizationSignedResponseAlg?: string;
    AuthorizationSignedResponseKeyID?: string;
    AuthorizationEncryptedResponseAlg?: string;
    AuthorizationEncryptedResponseEncryptionAlg?: string;

    IDTokenSignedResponseAlg?: string;
    IDTokenSignedResponseKeyID?: string;

    AccessTokenSignedResponseAlg?: string;
    AccessTokenSignedResponseKeyID?: string;

    IntrospectionSignedResponseAlg?: string;
    IntrospectionSignedResponseKeyID?: string;

    RequestObjectSigningAlg?: string;

    TokenEndpointAuthMethod?: string;
    TokenEnpointAuthSigningAlg?: string;

    RefreshFlowIgnoreOriginalGrantedScopes?: boolean;
    AllowMultipleAuthenticationMethods?: boolean;
    ClientCredentialsFlowAllowImplicitScope?: boolean;

    AuthorizationPolicy?: ClientAuthorizationPolicy;
    DefaultAuthorizationPolicy?: Policy;

    ConsentPolicy?: ClientConsentPolicy;
    RequestedAudienceMode?: string;

    ConsentMode?: string;
    ConsentPreConfiguredDuration?: number; //in seconds

    RequestURIs?: string[];
    JSONWebKeys?: string[]; //this is supposed to be abstracted behind JSONWebKeySet(type).JsonWebKey[]
    JSONWebKeysURI?: URL;
}

export enum ExistingScopes {
    openid = "openid",
    offline_access = "offline_access",
    groups = "groups",
    email = "email",
    profile = "profile",
    authelia_bearer_authz = "authelia.bearer.authz",
}

export interface ClientConsentPolicy {
    Mode: number;
    Duration: number; //seconds
}
export interface ClientAuthorizationPolicy {
    Name: string;
    Rules: ClientAuthorizationPolicyRule[];
}

// OIDC client auth policies are currently limited to policy and subject
export interface ClientAuthorizationPolicyRule {
    Subject: Subject;
    SetSubject(subjectString: string): void;
}

export enum SubjectPrefix {
    User = "user",
    Group = "group",
    OAuth2Client = "oauth2:client",
}

// OIDC Client Authorization Policy Subject
export interface Subject {
    prefix: SubjectPrefix;
    value: String[];
}

export type Policy = "one_factor" | "two_factor" | "deny";

export class ClientAuthorizationPolicyRuleImpl implements ClientAuthorizationPolicyRule {
    Subject: Subject;

    constructor(subjectString: string) {
        this.Subject = {
            prefix: SubjectPrefix.User,
            value: [],
        };
        this.SetSubject(subjectString);
    }

    SetSubject(subjectString: string): void {
        const parts = subjectString.split(":");

        if (parts.length === 3) {
            const prefix = parts[0].concat(":".concat(parts[1]));
            const values = parts.slice(2);
            if (prefix === "oauth2:client") {
                this.Subject = {
                    prefix: SubjectPrefix.OAuth2Client,
                    value: values,
                };
            }
        } else if (parts.length === 2) {
            const prefix = parts[0];
            const values = parts.slice(1);
            if (prefix === "user") {
                this.Subject = {
                    prefix: SubjectPrefix.User,
                    value: values,
                };
            } else if (prefix === "group") {
                this.Subject = {
                    prefix: SubjectPrefix.Group,
                    value: values,
                };
            } else {
                throw new Error("Invalid subject prefix. prefix: " + prefix + " value: " + values);
            }
        } else {
            throw new Error("Invalid subject format");
        }
    }
}

export type ResponseModeType =
    | ""
    | "form_post"
    | "query"
    | "fragment"
    | "form_post.jwt"
    | "query.jwt"
    | "fragment.jwt"
    | "jwt";

// TODO (Crowley723): define OIDC Provider
//begin OpenIDConnectProvider
export interface OpenIDConnectProvider {}

export interface IdentityProvidersOpenIDConnectLifespanGrants {
    AuthorizeCode: IdentityProvidersOpenIDConnectLifespanToken;
    Implicit: IdentityProvidersOpenIDConnectLifespanToken;
    ClientCredentials: IdentityProvidersOpenIDConnectLifespanToken;
    RefreshToken: IdentityProvidersOpenIDConnectLifespanToken;
    JWTBearer: IdentityProvidersOpenIDConnectLifespanToken;
}

export interface IdentityProvidersOpenIDConnectLifespanToken {
    AccessToken: number; //duration in seconds
    AuthorizeCode: number;
    IDToken: number;
    RefreshToken: number;
}
