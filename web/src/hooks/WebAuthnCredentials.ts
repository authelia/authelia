// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useRemoteCall } from "@hooks/RemoteCall";
import { getUserWebAuthnCredentials } from "@services/UserWebAuthnCredentials";

export function useUserWebAuthnCredentials() {
    return useRemoteCall(getUserWebAuthnCredentials);
}
