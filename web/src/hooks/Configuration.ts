// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useRemoteCall } from "@hooks/RemoteCall";
import { getConfiguration } from "@services/Configuration";

export function useConfiguration() {
    return useRemoteCall(getConfiguration, []);
}
