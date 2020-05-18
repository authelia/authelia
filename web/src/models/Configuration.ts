import { SecondFactorMethod } from "./Methods";

export interface Configuration {
    remember_me: boolean;
    reset_password: boolean;
    path: string;
}

export interface ExtendedConfiguration {
    available_methods: Set<SecondFactorMethod>;
    second_factor_enabled: boolean;
    totp_period: number;
}