import { OpenIDConnectClient } from "@models/OpenIDConnect";
import { OpenIDConnectClientConfigPath } from "@services/Api";
import { Get } from "@services/Client";

export async function getOpenIDConnectClients(): Promise<OpenIDConnectClient[]> {
    const res = await Get<OpenIDConnectClient[] | null>(OpenIDConnectClientConfigPath);

    if (res === null) {
        return [];
    }

    return res;
}
