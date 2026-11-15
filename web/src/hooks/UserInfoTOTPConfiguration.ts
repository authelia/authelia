// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useRemoteCall } from "@hooks/RemoteCall";
import {
    getUserInfoTOTPConfiguration,
    getUserInfoTOTPConfigurationOptional,
} from "@services/UserInfoTOTPConfiguration";

export function useUserInfoTOTPConfiguration() {
    return useRemoteCall(getUserInfoTOTPConfiguration);
}

export function useUserInfoTOTPConfigurationOptional() {
    return useRemoteCall(getUserInfoTOTPConfigurationOptional);
}
