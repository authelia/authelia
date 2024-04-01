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
