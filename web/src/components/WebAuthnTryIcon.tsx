import React, { useEffect } from "react";

import { Box, Button, useTheme } from "@mui/material";
import { useTranslation } from "react-i18next";

import FailureIcon from "@components/FailureIcon";
import FingerTouchIcon from "@components/FingerTouchIcon";
import LinearProgressBar from "@components/LinearProgressBar";
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
    const theme = useTheme();
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
            <LinearProgressBar value={timerPercent} height={theme.spacing(2)} />
        </IconWithContext>
    );

    const failure = (
        <IconWithContext
            icon={<FailureIcon />}
            className={props.webauthnTouchState === WebAuthnTouchState.Failure ? undefined : "hidden"}
        >
            <Button color="secondary" onClick={handleRetryClick} data-1p-ignore>
                {translate("Retry")}
            </Button>
        </IconWithContext>
    );

    return (
        <Box sx={{ minHeight: 101, display: "inline-block" }}>
            {touch}
            {failure}
        </Box>
    );
};

export default WebAuthnTryIcon;
