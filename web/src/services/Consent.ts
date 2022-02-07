import { ConsentPath } from "@services/Api";
import { Post, Get } from "@services/Client";

interface ConsentPostRequestBody {
    client_id: string;
    accept_or_reject: "accept" | "reject";
}

interface ConsentPostResponseBody {
    redirect_uri: string;
}

interface ConsentGetResponseBody {
    client_id: string;
    client_description: string;
    scopes: string[];
    audience: string[];
}

export function getRequestedScopes() {
    return Get<ConsentGetResponseBody>(ConsentPath);
}

export function acceptConsent(clientID: string) {
    const body: ConsentPostRequestBody = { client_id: clientID, accept_or_reject: "accept" };
    return Post<ConsentPostResponseBody>(ConsentPath, body);
}

export function rejectConsent(clientID: string) {
    const body: ConsentPostRequestBody = { client_id: clientID, accept_or_reject: "reject" };
    return Post<ConsentPostResponseBody>(ConsentPath, body);
}
