// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback } from "react";

export function useRedirector() {
    return useCallback((url: string) => {
        window.location.href = url;
    }, []);
}
