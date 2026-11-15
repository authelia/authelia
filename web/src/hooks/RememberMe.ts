// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useRef, useState } from "react";

export function useRememberMePrompt(enabled: boolean) {
    const [open, setOpen] = useState(false);

    const resolverRef = useRef<((_rememberMe: boolean | null) => void) | null>(null);

    const settle = useCallback((rememberMe: boolean | null) => {
        const resolve = resolverRef.current;

        resolverRef.current = null;

        setOpen(false);

        resolve?.(rememberMe);
    }, []);

    const handleChoice = useCallback((rememberMe: boolean) => settle(rememberMe), [settle]);

    const handleCancel = useCallback(() => settle(null), [settle]);

    useEffect(() => {
        return () => {
            resolverRef.current?.(null);
            resolverRef.current = null;
        };
    }, []);

    const prompt = useCallback(() => {
        if (!enabled) return Promise.resolve<boolean | null>(false);

        return new Promise<boolean | null>((resolve) => {
            resolverRef.current?.(null);
            resolverRef.current = resolve;

            setOpen(true);
        });
    }, [enabled]);

    return {
        dialogProps: {
            onCancel: handleCancel,
            onChoice: handleChoice,
            open,
        },
        prompt,
    };
}

export default useRememberMePrompt;
