// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useRemoteCall } from "@hooks/RemoteCall";
import { getUserSessionElevation } from "@services/UserSessionElevation";

export function useUserSessionElevation() {
    return useRemoteCall(getUserSessionElevation);
}
