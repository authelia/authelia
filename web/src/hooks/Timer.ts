import { useCallback, useEffect, useState } from "react";

const INTERVAL_MS = 100;

export function useTimer(timeoutMs: number): [number, () => void, () => void] {
    const [startTimestamp, setStartTimestamp] = useState<number | null>(null);
    const [percent, setPercent] = useState(0);

    const trigger = useCallback(() => {
        setPercent(0);
        setStartTimestamp(Date.now());
    }, []);

    const clear = useCallback(() => {
        setPercent(0);
        setStartTimestamp(null);
    }, []);

    useEffect(() => {
        if (startTimestamp === null) {
            return undefined;
        }

        const update = () => {
            const elapsedMs = Date.now() - startTimestamp;
            const completion = Math.min(100, (elapsedMs / timeoutMs) * 100);

            setPercent(completion);

            if (completion >= 100) {
                setStartTimestamp(null);
            }
        };

        const intervalId = setInterval(update, INTERVAL_MS);

        return () => {
            clearInterval(intervalId);
        };
    }, [startTimestamp, timeoutMs]);

    return [percent, trigger, clear];
}
