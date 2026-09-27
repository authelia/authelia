// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useSearchParams } from "react-router";

export function useQueryParam(queryParam: string) {
    const [query] = useSearchParams();
    const value = query.get(queryParam);
    return value !== "" ? (value as string) : undefined;
}
