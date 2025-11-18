import { Fragment } from "react";

import { Box, Theme } from "@mui/material";
import OtpInput from "react18-input-otp";
import { makeStyles } from "tss-react/mui";

import SuccessIcon from "@components/SuccessIcon";
import TimerIcon from "@components/TimerIcon";
import IconWithContext from "@views/LoginPortal/SecondFactor/IconWithContext";

export interface Props {
    passcode: string;
    state: State;

    digits: number;
    period: number;

    onChange: (_passcode: string) => void;
}

/* eslint-disable no-unused-vars */
export enum State {
    Idle = 1,
    InProgress = 2,
    Success = 3,
    Failure = 4,
    RateLimited = 5,
}

const OTPDial = function (props: Props) {
    const { classes, cx } = useStyles();

    return (
        <IconWithContext icon={<Icon state={props.state} period={props.period} />}>
            <Box component={"span"} className={classes.otpInput} id="otp-input">
                <OtpInput
                    shouldAutoFocus
                    onChange={props.onChange}
                    value={props.passcode}
                    numInputs={props.digits}
                    isDisabled={
                        props.state === State.InProgress ||
                        props.state === State.Success ||
                        props.state === State.RateLimited
                    }
                    isInputNum
                    hasErrored={props.state === State.Failure}
                    autoComplete="one-time-code"
                    inputStyle={cx(classes.otpDigitInput, props.state === State.Failure ? classes.inputError : "")}
                />
            </Box>
        </IconWithContext>
    );
};

interface IconProps {
    readonly state: State;
    readonly period: number;
}

function Icon(props: IconProps) {
    return (
        <Fragment>
            {props.state === State.Success ? (
                <SuccessIcon />
            ) : (
                <TimerIcon backgroundColor="#000" color="#FFFFFF" width={64} height={64} period={props.period} />
            )}
        </Fragment>
    );
}

const useStyles = makeStyles()((theme: Theme) => ({
    inputError: {
        border: "1px solid rgba(255, 2, 2, 0.95)",
    },
    otpDigitInput: {
        border: "1px solid rgba(0,0,0,0.3)",
        borderRadius: "5px",
        boxSizing: "content-box",
        fontSize: "1rem",
        marginLeft: theme.spacing(0.5),
        marginRight: theme.spacing(0.5),
        padding: theme.spacing() + " !important",
    },
    otpInput: {
        display: "inline-block",
        marginTop: theme.spacing(2),
    },
    register: {
        marginTop: theme.spacing(),
    },
    timeProgress: {},
}));

export default OTPDial;
