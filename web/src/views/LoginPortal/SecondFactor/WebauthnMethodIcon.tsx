import React, { Fragment } from "react";

import { Button, Theme, useTheme } from "@mui/material";
import { CSSProperties } from "@mui/styles";

import FailureIcon from "@components/FailureIcon";
import FingerTouchIcon from "@components/FingerTouchIcon";
import LinearProgressBar from "@components/LinearProgressBar";
import IconWithContext from "@views/LoginPortal/SecondFactor/IconWithContext";
import { State } from "@views/LoginPortal/SecondFactor/WebauthnMethod";

interface Props {
    state: State;

    timer: number;
    onRetryClick: () => void;
}

function WebauthnMethodIcon(props: Props) {
    const theme = useTheme();
    const style = useStyles(theme);

    const state = props.state as State;

    const progressBarStyle: CSSProperties = {
        marginTop: theme.spacing(),
    };

    const touch = (
        <IconWithContext
            icon={<FingerTouchIcon size={64} animated strong />}
            className={state === State.WaitTouch ? undefined : "hidden"}
        >
            <LinearProgressBar value={props.timer} sx={style.progressBar} />
        </IconWithContext>
    );

    const failure = (
        <IconWithContext icon={<FailureIcon />} className={state === State.Failure ? undefined : "hidden"}>
            <Button color="secondary" onClick={props.onRetryClick}>
                Retry
            </Button>
        </IconWithContext>
    );

    return (
        <Fragment>
            {touch}
            {failure}
        </Fragment>
    );
}

export default WebauthnMethodIcon;

const useStyles = (theme: Theme): { [key: string]: CSSProperties } => ({
    progressBar: {
        marginTop: theme.spacing(),
        height: theme.spacing(2),
    },
});
