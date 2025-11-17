import { useCallback, useState } from "react";

type PromisifiedFunction<Ret> = (...args: any) => Promise<Ret>;

export function useRemoteCall<Ret>(
    fn: PromisifiedFunction<Ret>,
): [Ret | undefined, () => void, boolean, Error | undefined] {
    const [data, setData] = useState(undefined as Ret | undefined);
    const [inProgress, setInProgress] = useState(false);
    const [error, setError] = useState(undefined as Error | undefined);

    const fnCallback = useCallback(() => fn(), [fn]);

    const triggerCallback = useCallback(() => {
        (async () => {
            try {
                setInProgress(true);
                const res = await fnCallback();
                setInProgress(false);
                setData(res);
            } catch (err) {
                console.error(err);
                setError(err as Error);
            }
        })();
    }, [setInProgress, setError, fnCallback]);

    return [data, triggerCallback, inProgress, error];
}
