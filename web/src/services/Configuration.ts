import { Configuration } from "@models/Configuration";
import { ConfigurationPath } from "@services/Api";
import { Get } from "@services/Client";
import { Method2FA, toSecondFactorMethod } from "@services/UserInfo";

interface ConfigurationPayload {
    available_methods: Method2FA[];
}

export async function getConfiguration(): Promise<Configuration> {
    const config = await Get<ConfigurationPayload>(ConfigurationPath);
    return { ...config, available_methods: new Set(config.available_methods.map(toSecondFactorMethod)) };
}
