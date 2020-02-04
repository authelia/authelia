import { Get } from "./Client";
import { ExtendedConfigurationPath, ConfigurationPath } from "./Api";
import { toEnum, Method2FA } from "./UserPreferences";
import { Configuration, ExtendedConfiguration } from "../models/Configuration";

export async function getConfiguration(): Promise<Configuration> {
    return Get<Configuration>(ConfigurationPath);
}

interface ExtendedConfigurationPayload {
    available_methods: Method2FA[];
    one_factor_default_policy: boolean;
}

export async function getExtendedConfiguration(): Promise<ExtendedConfiguration> {
    const config = await Get<ExtendedConfigurationPayload>(ExtendedConfigurationPath);
    return { ...config, available_methods: new Set(config.available_methods.map(toEnum)) };
}