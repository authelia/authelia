// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useRemoteCall } from "@hooks/RemoteCall";
import { getUserInfoTOTPConfiguration } from "@services/UserInfoTOTPConfiguration";

export function useUserInfoTOTPConfiguration() {
    return useRemoteCall(getUserInfoTOTPConfiguration, []);
}
