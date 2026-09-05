export interface OpenIDConnectProvider {
    id: string;
    name: string;
}

export interface OpenIDConnectStartResponse {
    authorization_url: string;
}

export interface OpenIDConnectLink {
    id: number;
    created_at: string;
    last_used_at?: string;
    provider: string;
    provider_name: string;
    issuer: string;
    subject: string;
    remote_username?: string;
}

export interface OpenIDConnectPendingLink {
    provider: string;
    provider_name: string;
    issuer: string;
    subject: string;
    remote_username?: string;
    display_name?: string;
    email?: string;
}

export interface OpenIDConnectLinks {
    links: OpenIDConnectLink[];
    pending?: OpenIDConnectPendingLink;
}
