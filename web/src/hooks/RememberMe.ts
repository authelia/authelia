// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useRef, useState } from "react";

export function useRememberMePrompt(enabled: boolean) {
    const [open, setOpen] = useState(false);

    const resolverRef = useRef<((_rememberMe: boolean) => void) | null>(null);

    const handleChoice = useCallback((rememberMe: boolean) => {
        const resolve = resolverRef.current;

        resolverRef.current = null;

        setOpen(false);

        resolve?.(rememberMe);
    }, []);

    const prompt = useCallback(() => {
        if (!enabled) return Promise.resolve(false);

        return new Promise<boolean>((resolve) => {
            // A prompt which is somehow still outstanding belongs to a superseded flow, so it declines rather than
            // being left dangling.
            resolverRef.current?.(false);
            resolverRef.current = resolve;

            setOpen(true);
        });
    }, [enabled]);

    return {
        dialogProps: {
            onChoice: handleChoice,
            open,
        },
        prompt,
    };
}

export default useRememberMePrompt;
