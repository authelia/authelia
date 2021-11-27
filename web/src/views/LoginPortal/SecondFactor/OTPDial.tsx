import React, { Fragment, ReactNode, useEffect } from "react";

import { makeStyles } from "@material-ui/core";
import classnames from "classnames";
import OtpInput from "react-otp-input";

import SuccessIcon from "@components/SuccessIcon";
import TimerIcon from "@components/TimerIcon";
import { useNotifications } from "@hooks/NotificationsContext";
import { useUserInfoTOTPConfiguration } from "@hooks/UserInfoTOTPConfiguration";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import IconWithContext from "@views/LoginPortal/SecondFactor/IconWithContext";
import { State } from "@views/LoginPortal/SecondFactor/OneTimePasswordMethod";

export interface Props {
    passcode: string;
    state: State;

    onChange: (passcode: string) => void;
}

const OTPDial = function (props: Props) {
    const style = useStyles();

    const [resp, fetch, , err] = useUserInfoTOTPConfiguration();

    const { createErrorNotification, resetNotification } = useNotifications();

    useEffect(() => {
        if (err) {
            console.error(`Failed to fetch TOTP configuration: ${err.message}`);
            createErrorNotification("Failed to obtain user One-Time Password Configuration.");
        }
    }, [resetNotification, createErrorNotification, err]);

    useEffect(() => {
        fetch();
    }, [fetch]);

    return (
        <div>
            {resp !== undefined && err === undefined ? (
                <IconWithContext icon={<Icon state={props.state} period={resp.period} />}>
                    <span className={style.otpInput} id="otp-input">
                        <OtpInput
                            shouldAutoFocus
                            onChange={props.onChange}
                            value={props.passcode}
                            numInputs={resp.digits}
                            isDisabled={props.state === State.InProgress || props.state === State.Success}
                            isInputNum
                            hasErrored={props.state === State.Failure}
                            inputStyle={classnames(
                                style.otpDigitInput,
                                props.state === State.Failure ? style.inputError : "",
                            )}
                        />
                    </span>
                </IconWithContext>
            ) : (
                <LoadingPage />
            )}
        </div>
    );
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
        boxSizing: "content-box",
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
