// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useRemoteCall } from "@hooks/RemoteCall";
import { getAllGroups } from "@services/GroupManagement";

export function useAllGroupsGET() {
    return useRemoteCall(getAllGroups);
}
