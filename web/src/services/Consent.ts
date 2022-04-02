import { ConsentPath } from "@services/Api";
import { Post, Get } from "@services/Client";

interface ConsentPostRequestBody {
    client_id: string;
    accept_or_reject: "accept" | "reject";
    pre_configure: boolean;
}

interface ConsentPostResponseBody {
    redirect_uri: string;
}

interface ConsentGetResponseBody {
    client_id: string;
    client_description: string;
    pre_configuration: boolean;
    scopes: string[];
    audience: string[];
}

export function getConsentRequest() {
    return Get<ConsentGetResponseBody>(ConsentPath);
}

export function acceptConsent(clientID: string, preConfigure: boolean = false) {
    const body: ConsentPostRequestBody = {
        client_id: clientID,
        accept_or_reject: "accept",
        pre_configure: preConfigure,
    };
    return Post<ConsentPostResponseBody>(ConsentPath, body);
}

export function rejectConsent(clientID: string) {
    const body: ConsentPostRequestBody = { client_id: clientID, accept_or_reject: "reject", pre_configure: false };
    return Post<ConsentPostResponseBody>(ConsentPath, body);
}
