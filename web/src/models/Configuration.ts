import { SecondFactorMethod } from "./Methods";

export interface Configuration {
    remember_me: boolean;
    reset_password: boolean;
}

export interface ExtendedConfiguration {
    available_methods: Set<SecondFactorMethod>;
    display_name: string;
    second_factor_enabled: boolean;
    totp_period: number;
}