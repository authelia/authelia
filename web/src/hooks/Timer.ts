import { useCallback, useEffect, useRef, useState } from "react";

export function useTimer(timeoutMs: number): [number, () => void, () => void] {
    const Interval = 100;
    const intervalRef = useRef<null | ReturnType<typeof setInterval>>(null);
    const timeoutMsRef = useRef(timeoutMs);
    const [percent, setPercent] = useState(0);

    useEffect(() => {
        timeoutMsRef.current = timeoutMs;
    }, [timeoutMs]);

    const stop = useCallback(() => {
        if (intervalRef.current !== null) {
            clearInterval(intervalRef.current);
            intervalRef.current = null;
        }
    }, []);

    const trigger = useCallback(() => {
        stop();
        setPercent(0);
        const startDate = new Date();
        intervalRef.current = setInterval(() => {
            const elapsedMs = new Date().getTime() - startDate.getTime();
            let p = (elapsedMs / timeoutMsRef.current) * 100.0;
            if (p >= 100) {
                p = 100;
                stop();
            }
            setPercent(p);
        }, Interval);
    }, [stop]);

    const clear = useCallback(() => {
        stop();
        setPercent(0);
    }, [stop]);

    useEffect(() => {
        return () => stop();
    }, [stop]);

    return [percent, trigger, clear];
}
