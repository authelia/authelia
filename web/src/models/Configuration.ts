import { SecondFactorMethod } from "@models/Methods";

export interface Configuration {
    available_methods: Set<SecondFactorMethod>;
    password_change_disabled: boolean;
    password_reset_disabled: boolean;
}

export interface SecuritySettingsConfiguration {
    disable: boolean;
}
