import { WebauthnDevice } from "@models/Webauthn";
import { WebauthnDevicesPath } from "@services/Api";
import { GetWithOptionalResponse } from "@services/Client";

// getWebauthnDevices returns the list of webauthn devices for the authenticated user.
export async function getWebauthnDevices(): Promise<WebauthnDevice[]> {
    return (await GetWithOptionalResponse<WebauthnDevice[]>(WebauthnDevicesPath)) || [];
}
