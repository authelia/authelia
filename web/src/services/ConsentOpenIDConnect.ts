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
    claims: string[] | null;
    essential_claims: string[] | null;
}

export function getConsentResponse(consentID: string) {
    return Get<ConsentGetResponseBody>(ConsentPath + "?id=" + consentID);
}

export function acceptConsent(preConfigure: boolean, clientID: string, claims: string[], consentID?: string) {
    const body: ConsentPostRequestBody = {
        id: consentID,
        client_id: clientID,
        consent: true,
        pre_configure: preConfigure,
        claims: claims,
    };
    return Post<ConsentPostResponseBody>(ConsentPath, body);
}

export function rejectConsent(clientID: string, consentID?: string) {
    const body: ConsentPostRequestBody = {
        id: consentID,
        client_id: clientID,
        consent: false,
        pre_configure: false,
    };
    return Post<ConsentPostResponseBody>(ConsentPath, body);
}

export function formatScope(scope: string, fallback: string): string {
    if (!scope.startsWith("scopes.") && scope !== "") {
        return scope;
    } else {
        return getScopeDescription(fallback);
    }
}

export function formatClaim(claim: string, fallback: string): string {
    if (!claim.startsWith("claims.") && claim !== "") {
        return claim;
    } else {
        return getClaimDescription(fallback);
    }
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
    claim = (claim.charAt(0).toUpperCase() + claim.slice(1)).replace("_verified", " (Verified)").replace("_", " ");

    for (let i = 0; i < claim.length; i++) {
        const j = i + 1;

        if (claim[i] === " " && j < claim.length) {
            claim.charAt(j).toUpperCase();
        }
    }
    return claim;
}
