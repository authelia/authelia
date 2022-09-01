import React, { useEffect, useState } from "react";

import PieChartIcon from "@components/PieChartIcon";

export interface Props {
    width: number;
    height: number;
    period: number;

    color?: string;
    backgroundColor?: string;
}

const TimerIcon = function (props: Props) {
    const radius = 31.6;
    const [timeProgress, setTimeProgress] = useState(0);

    useEffect(() => {
        // Get the current number of seconds to initialize timer.
        const initialValue = (((new Date().getTime() / 1000) % props.period) / props.period) * radius;
        setTimeProgress(initialValue);

        const interval = setInterval(() => {
            const value = (((new Date().getTime() / 1000) % props.period) / props.period) * radius;
            setTimeProgress(value);
        }, 100);
        return () => clearInterval(interval);
    }, [props]);

    return (
        <PieChartIcon
            width={props.width}
            height={props.height}
            progress={timeProgress}
            maxProgress={radius}
            backgroundColor={props.backgroundColor}
            color={props.color}
        />
    );
};

export default TimerIcon;
