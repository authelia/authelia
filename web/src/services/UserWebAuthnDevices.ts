import { WebAuthnDevice } from "@models/WebAuthn";
import { WebAuthnDevicesPath } from "@services/Api";
import { GetWithOptionalData } from "@services/Client";

// getWebAuthnDevices returns the list of webauthn devices for the authenticated user.
export async function getWebAuthnDevices(): Promise<WebAuthnDevice[] | null> {
    return GetWithOptionalData<WebAuthnDevice[] | null>(WebAuthnDevicesPath);
}
