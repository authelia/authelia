export interface OpenIDConnectClient {
    ID: string;
    Name: string;
    ClientType: ClientType; //aka Public

    //this should be ClientSecretDigest but I don't think we need to ever display it beyond creation/regeneration of secrets which can be done separately in plaintext
    //ClientSecret: string ;
    //RotatedClientSecrets?: string[];

    //begin advanced options
    SectorIdentifierURI?: string;
    RequirePushedAuthorizationRequest?: boolean;

    RequirePKCE?: boolean;
    RequirePKCEChallengeMethod?: boolean;
    PKCEChallengeMethod?: string;

    Audience: string[];
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

    AuthorizationPolicy: ClientAuthorizationPolicy;

    ConsentPolicy?: ClientConsentPolicy;
    RequestedAudienceMode?: string;

    ConsentMode?: string;
    ConsentPreConfiguredDuration?: number; //in seconds

    RequestURIs?: string[];
    JSONWebKeys?: string[]; //this is supposed to be abstracted behind JSONWebKeySet(type).JsonWebKey[]
    JSONWebKeysURI?: URL;
}

export enum ClientType {
    Public = "Public", //true
    Confidential = "Confidential", //false
}

export interface ClientConsentPolicy {
    Mode: number;
    Duration: number; //seconds
}
export interface ClientAuthorizationPolicy {
    Name: string;
    DefaultPolicy: number;
    Rules: ClientAuthorizationPolicyRule[];
}

// this is going to require quite a bit of additional infratructure to implement
// this is going to require ACLs to be defined
// TODO (Crowley723): ACL/Auth Policies need to be defined
export interface ClientAuthorizationPolicyRule {}

type ResponseModeType =
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
