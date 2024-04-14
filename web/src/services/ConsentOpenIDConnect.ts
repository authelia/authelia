import { ConsentPath } from "@services/Api";
import { Get, Post } from "@services/Client";

interface ConsentPostRequestBody {
    id?: string;
    client_id: string;
    consent: boolean;
    pre_configure: boolean;
    claims?: string[];
}

interface ConsentPostResponseBody {
    redirect_uri: string;
}

export interface ConsentGetResponseBody {
    client_id: string;
    client_description: string;
    scopes: string[];
    audience: string[];
    pre_configuration: boolean;
    claims: string[];
    essential_claims: string[];
}

export function getConsentResponse(consentID: string) {
    return Get<ConsentGetResponseBody>(ConsentPath + "?id=" + consentID);
}

export function acceptConsent(preConfigure: boolean, clientID: string, consentID: string | null, claims: string[]) {
    const body: ConsentPostRequestBody = {
        id: consentID === null ? undefined : consentID,
        client_id: clientID,
        consent: true,
        pre_configure: preConfigure,
        claims: claims,
    };
    return Post<ConsentPostResponseBody>(ConsentPath, body);
}

export function rejectConsent(clientID: string, consentID: string | null) {
    const body: ConsentPostRequestBody = {
        id: consentID === null ? undefined : consentID,
        client_id: clientID,
        consent: false,
        pre_configure: false,
    };
    return Post<ConsentPostResponseBody>(ConsentPath, body);
}

export function getScopeDescription(scope: string): string {
    switch (scope) {
        case "openid":
            return "Use OpenID to verify your identity";
        case "offline_access":
            return "Automatically refresh these permissions without user interaction";
        case "profile":
            return "Access your profile information";
        case "groups":
            return "Access your group membership";
        case "email":
            return "Access your email addresses";
        case "authelia.bearer.authz":
            return "Access protected resources logged in as you";
        default:
            return scope;
    }
}

export function getClaimDescription(claim: string): string {
    switch (claim) {
        case "name":
            return "Display Name";
        case "sub":
            return "Unique Identifier";
        case "zoneinfo":
            return "Timezone";
        case "locale":
            return "Locale / Language";
        case "updated_at":
            return "Information Updated Time";
        case "profile":
        case "website":
        case "picture":
            return `${setClaimCase(claim)} URL`;
        default:
            return setClaimCase(claim);
    }
}

function setClaimCase(claim: string): string {
    claim.charAt(0).toUpperCase();
    claim.replace("_verified", " (Verified)");
    claim.replace("_", " ");

    for (let i = 0; i < claim.length; i++) {
        const j = i + 1;

        if (claim[i] === " " && j < claim.length) {
            claim.charAt(j).toUpperCase();
        }
    }
    return claim;
}
