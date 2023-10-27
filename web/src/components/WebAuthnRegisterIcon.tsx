import React, { useEffect } from "react";

import { Box, Theme, useTheme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";

import FingerTouchIcon from "@components/FingerTouchIcon";
import LinearProgressBar from "@components/LinearProgressBar";
import { useTimer } from "@hooks/Timer";
import IconWithContext from "@views/LoginPortal/SecondFactor/IconWithContext";

interface Props {
    timeout: number;
}

const WebAuthnRegisterIcon = function (props: Props) {
    const theme = useTheme();
    const [timerPercent, triggerTimer] = useTimer(props.timeout);

    const styles = makeStyles((theme: Theme) => ({
        icon: {
            display: "inline-block",
        },
        progressBar: {
            marginTop: theme.spacing(),
        },
    }))();

    useEffect(() => {
        triggerTimer();
    }, [triggerTimer]);

    return (
        <Box className={styles.icon} sx={{ minHeight: 101 }}>
            <IconWithContext icon={<FingerTouchIcon size={64} animated strong />}>
                <LinearProgressBar value={timerPercent} className={styles.progressBar} height={theme.spacing(2)} />
            </IconWithContext>
        </Box>
    );
};

export default WebAuthnRegisterIcon;
