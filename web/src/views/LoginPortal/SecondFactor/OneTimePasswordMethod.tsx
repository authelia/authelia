import React, { Fragment, ReactNode, useCallback, useEffect, useRef, useState } from "react";

import { useRedirectionURL } from "@hooks/RedirectionURL";
import { useUserInfoTOTPConfiguration } from "@hooks/UserInfoTOTPConfiguration";
import { completeTOTPSignIn } from "@services/OneTimePassword";
import { AuthenticationLevel } from "@services/State";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import MethodContainer, { State as MethodContainerState } from "@views/LoginPortal/SecondFactor/MethodContainer";
import OTPDial from "@views/LoginPortal/SecondFactor/OTPDial";

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

    onRegisterClick: () => void;
    onSignInError: (err: Error) => void;
    onSignInSuccess: (redirectURL: string | undefined) => void;
}

const OneTimePasswordMethod = function (props: Props) {
    const [passcode, setPasscode] = useState("");
    const [state, setState] = useState(
        props.authenticationLevel === AuthenticationLevel.TwoFactor ? State.Success : State.Idle,
    );
    const redirectionURL = useRedirectionURL();

    const { onSignInSuccess, onSignInError } = props;
    const onSignInErrorCallback = useRef(onSignInError).current;
    const onSignInSuccessCallback = useRef(onSignInSuccess).current;

    const [resp, fetch, , err] = useUserInfoTOTPConfiguration();

    useEffect(() => {
        if (err) {
            console.error(`Failed to fetch TOTP configuration: ${err.message}`);
            onSignInErrorCallback(new Error("Failed to fetch user One-Time Password configuration"));
        }
    }, [onSignInErrorCallback, err]);

    useEffect(() => {
        fetch();
    }, [fetch]);

    const signInFunc = useCallback(async () => {
        if (!props.registered || props.authenticationLevel === AuthenticationLevel.TwoFactor) {
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
    }, [
        onSignInErrorCallback,
        onSignInSuccessCallback,
        passcode,
        redirectionURL,
        props.authenticationLevel,
        props.registered,
    ]);

    // Set successful state if user is already authenticated.
    useEffect(() => {
        if (props.authenticationLevel >= AuthenticationLevel.TwoFactor) {
            setState(State.Success);
        }
    }, [props.authenticationLevel, setState]);

    useEffect(() => {
        signInFunc();
    }, [signInFunc]);

    let methodState = MethodContainerState.METHOD;
    if (props.authenticationLevel === AuthenticationLevel.TwoFactor) {
        methodState = MethodContainerState.ALREADY_AUTHENTICATED;
    } else if (!props.registered) {
        methodState = MethodContainerState.NOT_REGISTERED;
    }

    return (
        <ComponentOrLoading ready={resp !== undefined}>
            {err !== undefined ? null : (
                <MethodContainer
                    id={props.id}
                    title="One-Time Password"
                    explanation="Enter one-time password"
                    registered={props.registered}
                    state={methodState}
                    onRegisterClick={props.onRegisterClick}
                >
                    <OTPDial passcode={passcode} onChange={setPasscode} state={state} period={resp.period} digits={resp.digits} />
                </MethodContainer>
                )
            }

        </ComponentOrLoading>
    );
};

export default OneTimePasswordMethod;

interface ComponentOrLoadingProps {
    ready: boolean;

    children: ReactNode;
}

function ComponentOrLoading(props: ComponentOrLoadingProps) {
    return (
        <Fragment>
            <div className={props.ready ? "hidden" : ""}>
                <LoadingPage />
            </div>
            {props.ready ? props.children : null}
        </Fragment>
    );
}