// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useRef, useState } from "react";

import { UserSessionElevation, getUserSessionElevation } from "@services/UserSessionElevation";

type Stage = "identity_verification" | "reauthentication" | "second_factor";

export interface ElevationFlowOptions<A extends string> {
    onElevated: (_action: A) => void;
    onCancelled: () => void;
    onRefreshError?: (_error: unknown) => void;
}

export function isElevated(elevation: undefined | UserSessionElevation) {
    return !!elevation && (elevation.elevated || elevation.skip_second_factor);
}

export function useElevationFlow<A extends string>(options: ElevationFlowOptions<A>) {
    const { onCancelled, onElevated, onRefreshError } = options;

    const [elevation, setElevation] = useState<UserSessionElevation>();
    const [pending, setPending] = useState<A>();
    const [stage, setStage] = useState<Stage>();
    const [opening, setOpening] = useState(false);

    // Invalidates outstanding fetches: bumped by reset() (and therefore cancel(), which calls
    // reset()) and by start(), so a fetch belonging to a superseded flow can recognize itself as
    // stale once it settles and avoid repopulating state or advancing the flow.
    const generationRef = useRef(0);

    const reset = useCallback(() => {
        generationRef.current += 1;

        setElevation(undefined);
        setPending(undefined);
        setStage(undefined);
        setOpening(false);
    }, []);

    const refresh = useCallback(async () => {
        const generation = generationRef.current;

        try {
            const result = await getUserSessionElevation();

            if (generationRef.current !== generation) {
                return undefined;
            }

            setElevation(result);

            return result;
        } catch (error) {
            if (generationRef.current !== generation) {
                console.error(error);

                return undefined;
            }

            throw error;
        }
    }, []);

    const advance = useCallback((next: Stage) => {
        setStage(next);
        setOpening(true);
    }, []);

    const complete = useCallback(() => {
        const action = pending;

        reset();

        if (action !== undefined) onElevated(action);
    }, [onElevated, pending, reset]);

    const cancel = useCallback(
        (dialog: string) => {
            console.warn(`${dialog} dialog close callback failed, it was likely cancelled by the user.`);

            reset();
            onCancelled();
        },
        [onCancelled, reset],
    );

    const handleRefreshError = useCallback(
        (error: unknown) => {
            console.error(error);
            onRefreshError?.(error);
        },
        [onRefreshError],
    );

    const completeIfReauthenticated = useCallback(() => {
        refresh()
            .then((result) => {
                if (result === undefined) {
                    return;
                }

                if (result.require_reauthentication === true) {
                    advance("reauthentication");

                    return;
                }

                complete();
            })
            .catch(handleRefreshError);
    }, [advance, complete, handleRefreshError, refresh]);

    const start = useCallback(
        (action: A) => {
            generationRef.current += 1;

            setPending(action);
            advance("reauthentication");

            refresh().catch(console.error);
        },
        [advance, refresh],
    );

    const handleOpened = useCallback(() => {
        setOpening(false);
    }, []);

    const handleReauthenticationClosed = useCallback(
        (ok: boolean, changed: boolean) => {
            if (!ok) {
                cancel("Reauthentication");

                return;
            }

            if (!changed) {
                advance("second_factor");

                return;
            }

            refresh()
                .then((result) => {
                    if (result === undefined) {
                        return;
                    }

                    advance("second_factor");
                })
                .catch(handleRefreshError);
        },
        [advance, cancel, handleRefreshError, refresh],
    );

    const handleSecondFactorClosed = useCallback(
        (ok: boolean, changed: boolean) => {
            if (!ok) {
                cancel("Second Factor");

                return;
            }

            const decide = (current: undefined | UserSessionElevation) => {
                if (isElevated(current)) {
                    completeIfReauthenticated();
                } else {
                    advance("identity_verification");
                }
            };

            if (!changed) {
                decide(elevation);

                return;
            }

            refresh()
                .then((result) => {
                    if (result === undefined) {
                        return;
                    }

                    decide(result);
                })
                .catch(handleRefreshError);
        },
        [advance, cancel, completeIfReauthenticated, elevation, handleRefreshError, refresh],
    );

    const handleIdentityVerificationClosed = useCallback(
        (ok: boolean) => {
            if (!ok) {
                cancel("Identity Verification");

                return;
            }

            completeIfReauthenticated();
        },
        [cancel, completeIfReauthenticated],
    );

    return {
        identityVerificationDialogProps: {
            elevation,
            handleClosed: handleIdentityVerificationClosed,
            handleOpened,
            opening: opening && stage === "identity_verification",
        },
        pending,
        reauthenticationDialogProps: {
            elevation,
            handleClosed: handleReauthenticationClosed,
            handleOpened,
            opening: opening && stage === "reauthentication",
        },
        reset,
        secondFactorDialogProps: {
            elevation,
            handleClosed: handleSecondFactorClosed,
            handleOpened,
            opening: opening && stage === "second_factor",
        },
        start,
    };
}
