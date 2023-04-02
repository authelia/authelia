// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { ChecksSafeRedirectionPath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";

interface SafeRedirectionResponse {
    ok: boolean;
}

export async function checkSafeRedirection(uri: string) {
    return PostWithOptionalResponse<SafeRedirectionResponse>(ChecksSafeRedirectionPath, { uri });
}
