import { Configuration } from "@models/Configuration";
import { ConfigurationLanguagesPath, ConfigurationPath } from "@services/Api";
import { Get } from "@services/Client";
import { toEnum, Method2FA } from "@services/UserInfo";

interface ConfigurationPayload {
    available_methods: Method2FA[];
}

export async function getConfiguration(): Promise<Configuration> {
    const config = await Get<ConfigurationPayload>(ConfigurationPath);
    return { ...config, available_methods: new Set(config.available_methods.map(toEnum)) };
}

interface ConfigurationLanguagesPayload {
    supported_languages: string[];
}

export function getConfigurationLanguages(): ConfigurationLanguagesPayload {
    const resp = Get<ConfigurationLanguagesPayload>(ConfigurationLanguagesPath);

    resp.then((result) => {
        return result;
    }).catch((err) => {
        throw err;
    });

    return { supported_languages: [] };
}
