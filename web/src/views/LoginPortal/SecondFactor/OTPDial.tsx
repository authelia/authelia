import { Fragment } from "react";

import { OTPField } from "@base-ui/react/otp-field";

import SuccessIcon from "@components/SuccessIcon";
import TimerIcon from "@components/TimerIcon";
import { cn } from "@utils/Styles";
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

const OTPDial = function (props: Props) {
    const disabled =
        props.state === State.InProgress || props.state === State.Success || props.state === State.RateLimited;

    return (
        <IconWithContext icon={<Icon state={props.state} period={props.period} />}>
            <span
                id="otp-input"
                className={cn(
                    "mt-4 inline-block",
                    "[&_input]:mx-1 [&_input]:box-border [&_input]:size-10 [&_input]:rounded-[5px] [&_input]:border [&_input]:border-black/30 [&_input]:text-center [&_input]:text-lg",
                    "[&_input:disabled]:cursor-not-allowed [&_input:disabled]:opacity-50",
                    props.state === State.Failure && "[&_input]:border-[rgba(255,2,2,0.95)]",
                )}
            >
                <OTPField.Root
                    length={props.digits}
                    value={props.passcode}
                    onValueChange={props.onChange}
                    disabled={disabled}
                    validationType="numeric"
                    inputMode="numeric"
                    autoComplete="one-time-code"
                >
                    {Array.from({ length: props.digits }, (_item, index) => (
                        <OTPField.Input key={index} autoFocus={index === 0} />
                    ))}
                </OTPField.Root>
            </span>
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
