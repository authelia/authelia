import { useEffect } from "react";

import { useTranslation } from "react-i18next";

import FailureIcon from "@components/FailureIcon";
import FingerTouchIcon from "@components/FingerTouchIcon";
import LinearProgressBar from "@components/LinearProgressBar";
import { Button } from "@components/UI/Button";
import { useTimer } from "@hooks/Timer";
import { WebAuthnTouchState } from "@models/WebAuthn";
import IconWithContext from "@views/LoginPortal/SecondFactor/IconWithContext";

interface Props {
    onRetryClick: () => void;
    webauthnTouchState: WebAuthnTouchState;
}

const WebAuthnTryIcon = function (props: Props) {
    const { t: translate } = useTranslation();
    const touchTimeout = 30;
    const [timerPercent, triggerTimer, clearTimer] = useTimer(touchTimeout * 1000 - 500);

    const handleRetryClick = () => {
        clearTimer();
        triggerTimer();
        props.onRetryClick();
    };

    useEffect(() => {
        triggerTimer();
    }, [triggerTimer]);

    const touch = (
        <IconWithContext
            icon={<FingerTouchIcon size={64} animated strong />}
            className={props.webauthnTouchState === WebAuthnTouchState.WaitTouch ? undefined : "hidden"}
        >
            <LinearProgressBar value={timerPercent} height={16} />
        </IconWithContext>
    );

    const failure = (
        <IconWithContext
            icon={<FailureIcon />}
            className={props.webauthnTouchState === WebAuthnTouchState.Failure ? undefined : "hidden"}
        >
            <Button
                variant="ghost"
                className="tracking-wide text-secondary hover:text-secondary"
                onClick={handleRetryClick}
            >
                {translate("Retry")}
            </Button>
        </IconWithContext>
    );

    return (
        <div className="inline-block" style={{ minHeight: 101 }}>
            {touch}
            {failure}
        </div>
    );
};

export default WebAuthnTryIcon;
