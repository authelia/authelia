// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useRef, useState } from "react";

// useRememberMePrompt drives a RememberMeDialog for a sign in flow which completes without the user having filled in
// the login form, and so without them having had the opportunity to tick the remember me checkbox on it. The prompt
// resolves to the choice the user makes, allowing the flow to await it at the point the session is about to be
// established rather than ahead of any user intent. When the feature is disabled the prompt resolves immediately, so
// callers do not have to special case it.
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
