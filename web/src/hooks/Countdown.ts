// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useEffect, useState } from "react";

function secondsUntil(expiresAt?: number) {
    if (expiresAt === undefined) return 0;

    return Math.max(0, Math.ceil(expiresAt - Date.now() / 1000));
}

/**
 * Returns the whole seconds remaining until a unix time in seconds, updated every second.
 */
export function useSecondsUntil(expiresAt?: number) {
    const [state, setState] = useState(() => ({ expiresAt, seconds: secondsUntil(expiresAt) }));

    if (state.expiresAt !== expiresAt) {
        setState({ expiresAt, seconds: secondsUntil(expiresAt) });
    }

    useEffect(() => {
        if (expiresAt === undefined) return;

        const interval = setInterval(() => {
            const seconds = secondsUntil(expiresAt);

            setState({ expiresAt, seconds });

            if (seconds === 0) clearInterval(interval);
        }, 1000);

        return () => clearInterval(interval);
    }, [expiresAt]);

    return state.seconds;
}
