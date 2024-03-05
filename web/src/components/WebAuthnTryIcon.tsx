import React, { useEffect } from "react";

import { Box, Button, Theme, useTheme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";

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
    const touchTimeout = 30;
    const theme = useTheme();
    const [timerPercent, triggerTimer, clearTimer] = useTimer(touchTimeout * 1000 - 500);

    const styles = makeStyles((theme: Theme) => ({
        icon: {
            display: "inline-block",
        },
        progressBar: {
            marginTop: theme.spacing(),
        },
    }))();

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
            <LinearProgressBar value={timerPercent} className={styles.progressBar} height={theme.spacing(2)} />
        </IconWithContext>
    );

    const failure = (
        <IconWithContext
            icon={<FailureIcon />}
            className={props.webauthnTouchState === WebAuthnTouchState.Failure ? undefined : "hidden"}
        >
            <Button color="secondary" onClick={handleRetryClick}>
                Retry
            </Button>
        </IconWithContext>
    );

    return (
        <Box className={styles.icon} sx={{ minHeight: 101 }}>
            {touch}
            {failure}
        </Box>
    );
};

export default WebAuthnTryIcon;
