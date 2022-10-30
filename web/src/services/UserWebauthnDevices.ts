import { WebauthnDevice } from "@root/models/Webauthn";
import { WebauthnDevicesPath } from "./Api";
import { Get } from "./Client";


// getWebauthnDevices returns the list of webauthn devices for the authenticated user.
export async function getWebauthnDevices(): Promise<WebauthnDevice[]> {
    return Get<WebauthnDevice[]>(WebauthnDevicesPath);
}