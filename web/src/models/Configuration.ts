import { SecondFactorMethod } from "./Methods";

export interface Configuration {
    available_methods: Set<SecondFactorMethod>;
    second_factor_enabled: boolean;
    totp_period: number;
}