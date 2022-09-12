import { useCallback, useEffect, useState } from "react";

export function useTimer(timeoutMs: number): [number, () => void, () => void] {
    const Interval = 100;
    const [startDate, setStartDate] = useState(undefined as Date | undefined);
    const [percent, setPercent] = useState(0);

    const trigger = useCallback(() => {
        setPercent(0);
        setStartDate(new Date());
    }, [setStartDate, setPercent]);

    const clear = useCallback(() => {
        setPercent(0);
        setStartDate(undefined);
    }, []);

    useEffect(() => {
        if (!startDate) {
            return;
        }

        const intervalNode = setInterval(() => {
            const elapsedMs = startDate ? new Date().getTime() - startDate.getTime() : 0;
            let p = (elapsedMs / timeoutMs) * 100.0;
            if (p >= 100) {
                p = 100;
                setStartDate(undefined);
            }
            setPercent(p);
        }, Interval);

        return () => clearInterval(intervalNode);
    }, [startDate, setPercent, setStartDate, timeoutMs]);

    return [percent, trigger, clear];
}
