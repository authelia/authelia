import { WebauthnDevice } from "@models/Webauthn";
import { WebAuthnDevicesPath } from "@services/Api";
import { GetWithOptionalData } from "@services/Client";

// getWebauthnDevices returns the list of webauthn devices for the authenticated user.
export async function getWebauthnDevices(): Promise<WebauthnDevice[] | null> {
    return GetWithOptionalData<WebauthnDevice[] | null>(WebAuthnDevicesPath);
}
