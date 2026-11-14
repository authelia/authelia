// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { SecondFactorMethod } from "@models/Methods";

export interface Configuration {
    app_authenticator_registration: boolean;
    available_methods: Set<SecondFactorMethod>;
    password_change_disabled: boolean;
    password_reset_disabled: boolean;
    recovery_codes_disabled: boolean;
}

export interface SecuritySettingsConfiguration {
    disable: boolean;
}
