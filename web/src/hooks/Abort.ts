import { useCallback, useEffect, useRef } from "react";

export function useAbortSignal(): () => AbortSignal {
    const ref = useRef<AbortController | null>(null);

    useEffect(() => {
        return () => {
            ref.current?.abort();
            ref.current = null;
        };
    }, []);

    return useCallback(() => {
        if (!ref.current || ref.current.signal.aborted) {
            ref.current = new AbortController();
        }
        return ref.current.signal;
    }, []);
}
