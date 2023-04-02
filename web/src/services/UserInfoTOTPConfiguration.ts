// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { UserInfoTOTPConfiguration } from "@models/UserInfoTOTPConfiguration";
import { UserInfoTOTPConfigurationPath } from "@services/Api";
import { Get } from "@services/Client";

export type TOTPDigits = 6 | 8;

export interface UserInfoTOTPConfigurationPayload {
    period: number;
    digits: TOTPDigits;
}

export async function getUserInfoTOTPConfiguration(): Promise<UserInfoTOTPConfiguration> {
    const res = await Get<UserInfoTOTPConfigurationPayload>(UserInfoTOTPConfigurationPath);
    return { ...res };
}
