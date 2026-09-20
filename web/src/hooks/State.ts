// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useRemoteCall } from "@hooks/RemoteCall";
import { getState } from "@services/State";

export function useAutheliaState() {
    return useRemoteCall(getState);
}
