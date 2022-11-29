import { WebauthnDevice } from "@models/Webauthn";
import { WebauthnDevicesPath } from "@services/Api";
import { Get } from "@services/Client";

// getWebauthnDevices returns the list of webauthn devices for the authenticated user.
export async function getWebauthnDevices(): Promise<WebauthnDevice[]> {
    return Get<WebauthnDevice[]>(WebauthnDevicesPath);
}
