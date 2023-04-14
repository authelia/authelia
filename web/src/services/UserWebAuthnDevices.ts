import { WebAuthnDevice } from "@models/WebAuthn";
import { WebAuthnDevicesPath } from "@services/Api";
import { GetWithOptionalData } from "@services/Client";

export async function getUserWebAuthnDevices(): Promise<WebAuthnDevice[]> {
    const res = await GetWithOptionalData<WebAuthnDevice[] | null>(WebAuthnDevicesPath);

    if (res === null) {
        return [];
    }

    return res;
}
