import { useEffect } from "react";

import { Box, useTheme } from "@mui/material";

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

    useEffect(() => {
        triggerTimer();
    }, [triggerTimer]);

    return (
        <Box sx={{ display: "inline-block", minHeight: 101 }}>
            <IconWithContext icon={<FingerTouchIcon size={64} animated strong />}>
                <LinearProgressBar value={timerPercent} height={theme.spacing(2)} />
            </IconWithContext>
        </Box>
    );
};

export default WebAuthnRegisterIcon;
