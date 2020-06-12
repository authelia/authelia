import { PostWithOptionalResponse, Get } from "./Client";
import { ConsentPath } from "./Api";
import { SignInResponse } from "./SignIn";

interface ConsentPostRequestBody {
    client_id: string;
}

interface ConsentGetResponseBody {
    client_id: string;
    scopes: string[];
}

export function getRequestedScopes() {
    return Get<ConsentGetResponseBody>(ConsentPath);
}

export function giveConsent(clientID: string) {
    const body: ConsentPostRequestBody = { client_id: clientID };
    return PostWithOptionalResponse<ConsentPostRequestBody>(ConsentPath, body);
}