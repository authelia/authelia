import { useEffect } from "react";

import FingerTouchIcon from "@components/FingerTouchIcon";
import LinearProgressBar from "@components/LinearProgressBar";
import { useTimer } from "@hooks/Timer";
import IconWithContext from "@views/LoginPortal/SecondFactor/IconWithContext";

interface Props {
    timeout: number;
}

const WebAuthnRegisterIcon = function (props: Props) {
    const [timerPercent, triggerTimer] = useTimer(props.timeout);

    useEffect(() => {
        triggerTimer();
    }, [triggerTimer]);

    return (
        <div className="inline-block" style={{ minHeight: 101 }}>
            <IconWithContext icon={<FingerTouchIcon size={64} animated strong />}>
                <LinearProgressBar value={timerPercent} height={16} />
            </IconWithContext>
        </div>
    );
};

export default WebAuthnRegisterIcon;
