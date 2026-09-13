// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

export interface ExternalIdentityProvider {
    id: string;
    name: string;
}

export interface ExternalIdentityStartResponse {
    authorization_url: string;
}

export interface ExternalIdentityLink {
    id: number;
    created_at: string;
    last_used_at?: string;
    type: string;
    provider: string;
    provider_name: string;
    issuer: string;
    subject: string;
    remote_username?: string;
    email?: string;
}

export interface ExternalIdentityPendingLink {
    provider: string;
    provider_name: string;
    issuer: string;
    subject: string;
    remote_username?: string;
    display_name?: string;
    email?: string;
}

export interface ExternalIdentityLinks {
    links: ExternalIdentityLink[];
    pending?: ExternalIdentityPendingLink;
}
