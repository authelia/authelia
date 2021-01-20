import { Configuration } from "../models/Configuration";
import { ConfigurationPath } from "./Api";
import { Get } from "./Client";
import { toEnum, Method2FA } from "./UserPreferences";

interface ConfigurationPayload {
    available_methods: Method2FA[];
    second_factor_enabled: boolean;
    totp_period: number;
}

export async function getConfiguration(): Promise<Configuration> {
    const config = await Get<ConfigurationPayload>(ConfigurationPath);
    return { ...config, available_methods: new Set(config.available_methods.map(toEnum)) };
}
