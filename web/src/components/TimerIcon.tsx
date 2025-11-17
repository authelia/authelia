import { useEffect, useState } from "react";

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
    const [timeProgress, setTimeProgress] = useState(() => {
        return (((Date.now() / 1000) % props.period) / props.period) * radius;
    });

    useEffect(() => {
        const interval = setInterval(() => {
            const value = (((Date.now() / 1000) % props.period) / props.period) * radius;
            setTimeProgress(value);
        }, 100);
        return () => clearInterval(interval);
    }, [props.period]);

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
