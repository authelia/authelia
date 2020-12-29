import React, { Fragment } from "react";

import { makeStyles } from "@material-ui/core";
import classnames from "classnames";
import OtpInput from "react-otp-input";

import SuccessIcon from "../../../components/SuccessIcon";
import TimerIcon from "../../../components/TimerIcon";
import IconWithContext from "./IconWithContext";
import { State } from "./OneTimePasswordMethod";

export interface Props {
    passcode: string;
    state: State;
    period: number;

    onChange: (passcode: string) => void;
}

const OTPDial = function (props: Props) {
    const style = useStyles();
    const dial = (
        <span className={style.otpInput} id="otp-input">
            <OtpInput
                shouldAutoFocus
                onChange={props.onChange}
                value={props.passcode}
                numInputs={6}
                isDisabled={props.state === State.InProgress || props.state === State.Success}
                hasErrored={props.state === State.Failure}
                inputStyle={classnames(style.otpDigitInput, props.state === State.Failure ? style.inputError : "")}
            />
        </span>
    );

    return <IconWithContext icon={<Icon state={props.state} period={props.period} />} context={dial} />;
};

export default OTPDial;

const useStyles = makeStyles((theme) => ({
    timeProgress: {},
    register: {
        marginTop: theme.spacing(),
    },
    otpInput: {
        display: "inline-block",
        marginTop: theme.spacing(2),
    },
    otpDigitInput: {
        padding: theme.spacing(),
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
