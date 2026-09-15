// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { ChecksSafePostLogoutRedirectionPath, ChecksSafeRedirectionPath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";

interface SafeRedirectionResponse {
    ok: boolean;
}

export async function checkSafeRedirection(uri: string) {
    return PostWithOptionalResponse<SafeRedirectionResponse>(ChecksSafeRedirectionPath, { uri });
}

export async function checkSafePostLogoutRedirection(uri: string) {
    return PostWithOptionalResponse<SafeRedirectionResponse>(ChecksSafePostLogoutRedirectionPath, { uri });
}
