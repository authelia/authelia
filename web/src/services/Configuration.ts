import { Configuration } from "@models/Configuration";
import { ConfigurationPath } from "@services/Api";
import { Get } from "@services/Client";
import { Method2FA, toSecondFactorMethod } from "@services/UserInfo";

interface ConfigurationPayload {
    available_methods: Method2FA[];
    password_change_disabled: boolean;
    password_reset_disabled: boolean;
}

export async function getConfiguration(): Promise<Configuration> {
    const config = await Get<ConfigurationPayload>(ConfigurationPath);
    return {
        ...config,
        available_methods: new Set(config.available_methods.map(toSecondFactorMethod)),
        password_change_disabled: config.password_change_disabled,
        password_reset_disabled: config.password_reset_disabled,
    };
}
