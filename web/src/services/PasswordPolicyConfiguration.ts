import { PasswordPolicyConfiguration, PasswordPolicyMode } from "@models/PasswordPolicy";
import { PasswordPolicyConfigurationPath } from "@services/Api";
import { Get } from "@services/Client";

interface PasswordPolicyConfigurationPayload {
    mode: ModePasswordPolicy;
    min_length: number;
    max_length: number;
    min_score: number;
    require_uppercase: boolean;
    require_lowercase: boolean;
    require_number: boolean;
    require_special: boolean;
}

export type ModePasswordPolicy = "disabled" | "standard" | "zxcvbn";

export function toEnum(method: ModePasswordPolicy): PasswordPolicyMode {
    switch (method) {
        case "disabled":
            return PasswordPolicyMode.Disabled;
        case "standard":
            return PasswordPolicyMode.Standard;
        case "zxcvbn":
            return PasswordPolicyMode.ZXCVBN;
    }
}

export async function getPasswordPolicyConfiguration(): Promise<PasswordPolicyConfiguration> {
    const config = await Get<PasswordPolicyConfigurationPayload>(PasswordPolicyConfigurationPath);

    return { ...config, mode: toEnum(config.mode) };
}
