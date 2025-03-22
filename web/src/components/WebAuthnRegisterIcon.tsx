import React, { useEffect } from "react";

import { Box, Theme, useTheme } from "@mui/material";
import { makeStyles } from "tss-react/mui";

import FingerTouchIcon from "@components/FingerTouchIcon";
import LinearProgressBar from "@components/LinearProgressBar";
import { useTimer } from "@hooks/Timer";
import IconWithContext from "@views/LoginPortal/SecondFactor/IconWithContext";

interface Props {
    timeout: number;
}

const WebAuthnRegisterIcon = function (props: Props) {
    const { classes } = useStyles();

    const theme = useTheme();
    const [timerPercent, triggerTimer] = useTimer(props.timeout);

    useEffect(() => {
        triggerTimer();
    }, [triggerTimer]);

    return (
        <Box className={classes.icon} sx={{ minHeight: 101 }}>
            <IconWithContext icon={<FingerTouchIcon size={64} animated strong />}>
                <LinearProgressBar value={timerPercent} height={theme.spacing(2)} />
            </IconWithContext>
        </Box>
    );
};

const useStyles = makeStyles()((theme: Theme) => ({
    icon: {
        display: "inline-block",
    },
}));

export default WebAuthnRegisterIcon;
