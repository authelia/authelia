// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useRemoteCall } from "@hooks/RemoteCall";
import { getAppAuthenticatorDevices } from "@services/AppAuthenticator";

export function useAppAuthenticatorDevices() {
    return useRemoteCall(getAppAuthenticatorDevices);
}
