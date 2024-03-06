import React, { Fragment } from "react";

import { Theme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import classnames from "classnames";
import OtpInput from "react18-input-otp";

import SuccessIcon from "@components/SuccessIcon";
import TimerIcon from "@components/TimerIcon";
import IconWithContext from "@views/LoginPortal/SecondFactor/IconWithContext";
import { State } from "@views/LoginPortal/SecondFactor/OneTimePasswordMethod";

export interface Props {
    passcode: string;
    state: State;

    digits: number;
    period: number;

    onChange: (passcode: string) => void;
}

const OTPDial = function (props: Props) {
    const styles = useStyles();

    return (
        <IconWithContext icon={<Icon state={props.state} period={props.period} />}>
            <span className={styles.otpInput} id="otp-input">
                <OtpInput
                    shouldAutoFocus
                    onChange={props.onChange}
                    value={props.passcode}
                    numInputs={props.digits}
                    isDisabled={props.state === State.InProgress || props.state === State.Success}
                    isInputNum
                    hasErrored={props.state === State.Failure}
                    autoComplete="one-time-code"
                    inputStyle={classnames(
                        styles.otpDigitInput,
                        props.state === State.Failure ? styles.inputError : "",
                    )}
                />
            </span>
        </IconWithContext>
    );
};

const useStyles = makeStyles((theme: Theme) => ({
    timeProgress: {},
    register: {
        marginTop: theme.spacing(),
    },
    otpInput: {
        display: "inline-block",
        marginTop: theme.spacing(2),
    },
    otpDigitInput: {
        boxSizing: "content-box",
        padding: theme.spacing() + " !important",
        marginLeft: theme.spacing(0.5),
        marginRight: theme.spacing(0.5),
        fontSize: "1rem",
        borderRadius: "5px",
        border: "1px solid rgba(0,0,0,0.3)",
    },
    inputError: {
        border: "1px solid rgba(255, 2, 2, 0.95)",
    },
}));

interface IconProps {
    state: State;
    period: number;
}

function Icon(props: IconProps) {
    return (
        <Fragment>
            {props.state !== State.Success ? (
                <TimerIcon backgroundColor="#000" color="#FFFFFF" width={64} height={64} period={props.period} />
            ) : null}
            {props.state === State.Success ? <SuccessIcon /> : null}
        </Fragment>
    );
}

export default OTPDial;
