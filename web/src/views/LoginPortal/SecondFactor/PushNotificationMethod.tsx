import React, { useEffect, useCallback, useState, ReactNode } from "react";

import { Button, makeStyles } from "@material-ui/core";

import FailureIcon from "../../../components/FailureIcon";
import PushNotificationIcon from "../../../components/PushNotificationIcon";
import SuccessIcon from "../../../components/SuccessIcon";
import { useIsMountedRef } from "../../../hooks/Mounted";
import { useRedirectionURL } from "../../../hooks/RedirectionURL";
import { completePushNotificationSignIn } from "../../../services/PushNotification";
import { AuthenticationLevel } from "../../../services/State";
import MethodContainer, { State as MethodContainerState } from "./MethodContainer";

export enum State {
    SignInInProgress = 1,
    Success = 2,
    Failure = 3,
}

export interface Props {
    id: string;
    authenticationLevel: AuthenticationLevel;

    onSignInError: (err: Error) => void;
    onSignInSuccess: (redirectURL: string | undefined) => void;
}

const PushNotificationMethod = function (props: Props) {
    const style = useStyles();
    const [state, setState] = useState(State.SignInInProgress);
    const redirectionURL = useRedirectionURL();
    const mounted = useIsMountedRef();

    const { onSignInSuccess, onSignInError } = props;
    /* eslint-disable react-hooks/exhaustive-deps */
    const onSignInErrorCallback = useCallback(onSignInError, []);
    const onSignInSuccessCallback = useCallback(onSignInSuccess, []);
    /* eslint-enable react-hooks/exhaustive-deps */

    const signInFunc = useCallback(async () => {
        if (props.authenticationLevel === AuthenticationLevel.TwoFactor) {
            return;
        }

        try {
            setState(State.SignInInProgress);
            const res = await completePushNotificationSignIn(redirectionURL);
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;

            setState(State.Success);
            setTimeout(() => {
                if (!mounted.current) return;
                onSignInSuccessCallback(res ? res.redirect : undefined);
            }, 1500);
        } catch (err) {
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;

            console.error(err);
            onSignInErrorCallback(new Error("There was an issue completing sign in process"));
            setState(State.Failure);
        }
    }, [onSignInErrorCallback, onSignInSuccessCallback, setState, redirectionURL, mounted, props.authenticationLevel]);

    useEffect(() => {
        signInFunc();
    }, [signInFunc]);

    // Set successful state if user is already authenticated.
    useEffect(() => {
        if (props.authenticationLevel >= AuthenticationLevel.TwoFactor) {
            setState(State.Success);
        }
    }, [props.authenticationLevel, setState]);

    let icon: ReactNode;
    switch (state) {
        case State.SignInInProgress:
            icon = <PushNotificationIcon width={64} height={64} animated />;
            break;
        case State.Success:
            icon = <SuccessIcon />;
            break;
        case State.Failure:
            icon = <FailureIcon />;
    }

    let methodState = MethodContainerState.METHOD;
    if (props.authenticationLevel === AuthenticationLevel.TwoFactor) {
        methodState = MethodContainerState.ALREADY_AUTHENTICATED;
    }

    return (
        <MethodContainer
            id={props.id}
            title="Push Notification"
            explanation="A notification has been sent to your smartphone"
            state={methodState}
        >
            <div className={style.icon}>{icon}</div>
            <div className={state !== State.Failure ? "hidden" : ""}>
                <Button color="secondary" onClick={signInFunc}>
                    Retry
                </Button>
            </div>
        </MethodContainer>
    );
};

export default PushNotificationMethod;

const useStyles = makeStyles((theme) => ({
    icon: {
        width: "64px",
        height: "64px",
        display: "inline-block",
    },
}));
