import { ConsentPath } from "@services/Api";
import { Post, Get } from "@services/Client";

interface ConsentPostRequestBody {
    client_id: string;
    consent_id?: string;
    consent: boolean;
    pre_configure: boolean;
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
}

export function getConsentResponse(consentID: string) {
    return Get<ConsentGetResponseBody>(ConsentPath + "?consent_id=" + consentID);
}

export function acceptConsent(preConfigure: boolean, clientID: string, consentID?: string) {
    const body: ConsentPostRequestBody = {
        client_id: clientID,
        consent_id: consentID,
        consent: true,
        pre_configure: preConfigure,
    };
    return Post<ConsentPostResponseBody>(ConsentPath, body);
}

export function rejectConsent(clientID: string, consentID?: string) {
    const body: ConsentPostRequestBody = {
        client_id: clientID,
        consent_id: consentID,
        consent: false,
        pre_configure: false,
    };
    return Post<ConsentPostResponseBody>(ConsentPath, body);
}
