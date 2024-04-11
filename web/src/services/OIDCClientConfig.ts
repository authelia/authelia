import { OpenIDConnectClient } from "@models/OpenIDConnect";
import { OpenIDConnectClientConfigPath } from "@services/Api";
import { Get } from "@services/Client";

export async function getOpenIDConnectClients(): Promise<OpenIDConnectClient[]> {
    try {
        const res = await Get<OpenIDConnectClient[] | null>(OpenIDConnectClientConfigPath);
        console.log("Error fetching OpenIDConnect clients:");
        if (res === null) {
            return [];
        }

        return res;
    } catch (error) {
        console.error("Error fetching OpenIDConnect clients:", error);
        return []; // Or handle the error in an appropriate way
    }
}
