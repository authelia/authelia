import { SecondFactorMethod } from "./Methods";

export interface Configuration {
    ga_tracking_id: string;
    remember_me_enabled: boolean;
}

export interface ExtendedConfiguration {
    available_methods: Set<SecondFactorMethod>;
    second_factor_enabled: boolean;
    totp_period: number;
}