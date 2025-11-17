import { WebAuthnCredential } from "@models/WebAuthn";
import { WebAuthnCredentialsPath } from "@services/Api";
import { GetWithOptionalData } from "@services/Client";

export async function getUserWebAuthnCredentials(): Promise<WebAuthnCredential[]> {
    const res = await GetWithOptionalData<null | WebAuthnCredential[]>(WebAuthnCredentialsPath);

    if (res === null) {
        return [];
    }

    return res;
}
