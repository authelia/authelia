// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useState } from "react";

type PromisifiedFunction<Ret> = (..._args: any) => Promise<Ret>;

export interface RemoteCallOptions {
    resetDataOnCall?: boolean;
}

export function useRemoteCall<Ret>(
    fn: PromisifiedFunction<Ret>,
    options?: RemoteCallOptions,
): [Ret | undefined, () => void, boolean, Error | undefined] {
    const [data, setData] = useState(undefined as Ret | undefined);
    const [inProgress, setInProgress] = useState(false);
    const [error, setError] = useState(undefined as Error | undefined);

    const resetDataOnCall = options?.resetDataOnCall ?? false;

    const fnCallback = useCallback(() => fn(), [fn]);

    const triggerCallback = useCallback(() => {
        (async () => {
            try {
                setInProgress(true);
                setError(undefined);

                if (resetDataOnCall) {
                    setData(undefined);
                }

                const res = await fnCallback();
                setData(res);
            } catch (err) {
                console.error(err);
                setError(err as Error);
            } finally {
                setInProgress(false);
            }
        })();
    }, [setInProgress, setError, fnCallback, resetDataOnCall]);

    return [data, triggerCallback, inProgress, error];
}
