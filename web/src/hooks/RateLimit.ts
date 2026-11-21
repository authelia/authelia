// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useEffectEvent, useState } from "react";

const MaximumTimeoutDelay = 2_147_483_647;

function secondsUntil(deadline: number) {
    return Math.max(0, Math.ceil((deadline - Date.now()) / 1000));
}

export const useRateLimit = (onExpire?: () => void) => {
    const [deadline, setDeadline] = useState<null | number>(null);
    const [secondsRemaining, setSecondsRemaining] = useState(0);

    const handleExpire = useEffectEvent(() => {
        setDeadline(null);
        setSecondsRemaining(0);
        onExpire?.();
    });

    useEffect(() => {
        if (deadline === null) return;

        let interval: null | ReturnType<typeof setInterval> = null;
        let timeout: null | ReturnType<typeof setTimeout> = null;

        const stop = () => {
            if (interval !== null) clearInterval(interval);
            if (timeout !== null) clearTimeout(timeout);

            interval = null;
            timeout = null;
        };

        const expire = () => {
            stop();
            handleExpire();
        };

        interval = setInterval(() => {
            const seconds = secondsUntil(deadline);

            if (seconds === 0) {
                expire();
            } else {
                setSecondsRemaining(seconds);
            }
        }, 1000);

        const delay = Math.max(0, deadline - Date.now());

        if (delay <= MaximumTimeoutDelay) timeout = setTimeout(expire, delay);

        return stop;
    }, [deadline]);

    const clear = useCallback(() => {
        setDeadline(null);
        setSecondsRemaining(0);
    }, []);

    const trigger = useCallback((retryAfter: number) => {
        setDeadline(Date.now() + retryAfter * 1000);
        setSecondsRemaining(Math.ceil(retryAfter));
    }, []);

    return { clear, limited: deadline !== null, secondsRemaining, trigger };
};

export default useRateLimit;
