import React, { useState, useEffect, useCallback } from "react";
import MethodContainer, { State as MethodContainerState } from "./MethodContainer";
import OTPDial from "./OTPDial";
import { completeTOTPSignIn } from "../../../services/OneTimePassword";
import { useRedirectionURL } from "../../../hooks/RedirectionURL";
import { AuthenticationLevel } from "../../../services/State";

export enum State {
    Idle = 1,
    InProgress = 2,
    Success = 3,
    Failure = 4,
}

export interface Props {
    id: string;
    authenticationLevel: AuthenticationLevel;
    registered: boolean;
    totp_period: number

    onRegisterClick: () => void;
    onSignInError: (err: Error) => void;
    onSignInSuccess: (redirectURL: string | undefined) => void;
}

export default function (props: Props) {
    const [passcode, setPasscode] = useState("");
    const [state, setState] = useState(props.authenticationLevel === AuthenticationLevel.TwoFactor
        ? State.Success
        : State.Idle);
    const redirectionURL = useRedirectionURL();

    const { onSignInSuccess, onSignInError } = props;
    const onSignInErrorCallback = useCallback(onSignInError, []);
    const onSignInSuccessCallback = useCallback(onSignInSuccess, []);

    const signInFunc = useCallback(async () => {
        if (props.authenticationLevel === AuthenticationLevel.TwoFactor) {
            return;
        }

        const passcodeStr = `${passcode}`;

        if (!passcode || passcodeStr.length !== 6) {
            return;
        }

        try {
            setState(State.InProgress);
            const res = await completeTOTPSignIn(passcodeStr, redirectionURL);
            setState(State.Success);
            onSignInSuccessCallback(res ? res.redirect : undefined);
        } catch (err) {
            console.error(err);
            onSignInErrorCallback(new Error("The one-time password might be wrong"));
            setState(State.Failure);
        }
        setPasscode("");
    }, [passcode, onSignInErrorCallback, onSignInSuccessCallback, redirectionURL, props.authenticationLevel]);

    // Set successful state if user is already authenticated.
    useEffect(() => {
        if (props.authenticationLevel >= AuthenticationLevel.TwoFactor) {
            setState(State.Success);
        }
    }, [props.authenticationLevel, setState]);

    useEffect(() => { signInFunc() }, [signInFunc]);

    let methodState = MethodContainerState.METHOD;
    if (props.authenticationLevel === AuthenticationLevel.TwoFactor) {
        methodState = MethodContainerState.ALREADY_AUTHENTICATED;
    } else if (!props.registered) {
        methodState = MethodContainerState.NOT_REGISTERED;
    }

    return (
        <MethodContainer
            id={props.id}
            title="One-Time Password"
            explanation="Enter one-time password"
            state={methodState}
            onRegisterClick={props.onRegisterClick}>
            <OTPDial
                passcode={passcode}
                onChange={setPasscode}
                state={state}
                period={props.totp_period} />
        </MethodContainer>
    )
}