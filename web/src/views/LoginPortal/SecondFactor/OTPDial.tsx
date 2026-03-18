import { Fragment } from "react";

import { styled } from "@mui/material";
import _OtpInputModule from "react18-input-otp";

const OtpInput = (_OtpInputModule as { default?: typeof _OtpInputModule }).default ?? _OtpInputModule;

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

export enum State {
    Idle = 1,
    InProgress = 2,
    Success = 3,
    Failure = 4,
    RateLimited = 5,
}

const StyledOtpInputContainer = styled("span")(({ theme }) => ({
    "& input": {
        border: "1px solid rgba(0,0,0,0.3)",
        borderRadius: "5px",
        boxSizing: "content-box",
        fontSize: "1rem",
        marginLeft: theme.spacing(0.5),
        marginRight: theme.spacing(0.5),
        padding: `${theme.spacing()} !important`,
    },
    display: "inline-block",
    marginTop: theme.spacing(2),
}));

const OTPDial = function (props: Props) {
    return (
        <IconWithContext icon={<Icon state={props.state} period={props.period} />}>
            <StyledOtpInputContainer
                id="otp-input"
                sx={
                    props.state === State.Failure
                        ? { "& input": { border: "1px solid rgba(255, 2, 2, 0.95)" } }
                        : undefined
                }
            >
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
                />
            </StyledOtpInputContainer>
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

export default OTPDial;
