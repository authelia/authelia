// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useRemoteCall } from "@hooks/RemoteCall";
import { getUserInfo, postUserInfo } from "@services/UserInfo";

export function useUserInfoPOST() {
    return useRemoteCall(postUserInfo);
}

export function useUserInfoGET() {
    return useRemoteCall(getUserInfo);
}
