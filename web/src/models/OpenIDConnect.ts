export enum ClientType {
    Public = "Public",
    Confidential = "Confidential",
}

export interface OpenIDConnectClient {
    ID: string;
    Name: string;
    ClientType: ClientType;
    RedirectURIs: string[];
    Audience: string[];
    Scopes: string[];
    AuthorizationPolicy: ClientAuthorizationPolicy;

    //begin advanced options
    
}

export interface ClientAuthorizationPolicy {
    Name: string;
    DefaultPolicy: number;
    Rules: ClientAuthorizationPolicyRule[];
}

// this is going to require quite a bit of additional infratructure to implement
// this is going to require ACLs to be defined
// or maybe not, im not entirely sure.
export interface ClientAuthorizationPolicyRule {}

export interface OpenIDConnectProvider {}
