import React, { useState, useEffect } from "react";
import PieChartIcon from "./PieChartIcon";

export interface Props {
    width: number;
    height: number;
    period: number;

    color?: string;
    backgroundColor?: string;
}

export default function (props: Props) {
    const maxTimeProgress = 1000;
    const [timeProgress, setTimeProgress] = useState(0);

    useEffect(() => {
        // Get the current number of seconds to initialize timer.
        const periodMs = props.period * 1000;
        const initialValue = Math.floor((new Date().getSeconds() % props.period) / props.period * maxTimeProgress);
        setTimeProgress(initialValue);

        const interval = setInterval(() => {
            const ms = new Date().getSeconds() * 1000.0 + new Date().getMilliseconds();
            const value = (ms % periodMs) / periodMs * maxTimeProgress;
            setTimeProgress(value);
        }, 100);
        return () => clearInterval(interval);
    }, [props]);

    return (
        <PieChartIcon width={props.width} height={props.height}
            maxProgress={maxTimeProgress}
            progress={timeProgress}
            backgroundColor={props.backgroundColor} color={props.color} />
    )
}

