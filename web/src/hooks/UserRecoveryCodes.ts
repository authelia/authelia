// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useRemoteCall } from "@hooks/RemoteCall";
import { generateRecoveryCodes, getRecoveryCodesStatus } from "@services/UserRecoveryCodes";

export function useRecoveryCodesStatus() {
    return useRemoteCall(getRecoveryCodesStatus);
}

export function useGenerateRecoveryCodes() {
    return useRemoteCall(generateRecoveryCodes);
}
