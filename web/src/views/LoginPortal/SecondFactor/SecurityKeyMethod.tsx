import React, { useCallback, useEffect, useState, Fragment } from "react";
import MethodContainer from "./MethodContainer";
import { makeStyles, Button, useTheme } from "@material-ui/core";
import { initiateU2FSignin, completeU2FSignin } from "../../../services/SecurityKey";
import u2fApi from "u2f-api";
import { useRedirectionURL } from "../../../hooks/RedirectionURL";
import { useIsMountedRef } from "../../../hooks/Mounted";
import { useTimer } from "../../../hooks/Timer";
import LinearProgressBar from "../../../components/LinearProgressBar";
import FingerTouchIcon from "../../../components/FingerTouchIcon";
import SuccessIcon from "../../../components/SuccessIcon";
import FailureIcon from "../../../components/FailureIcon";
import IconWithContext from "./IconWithContext";
import { CSSProperties } from "@material-ui/styles";
import { AuthenticationLevel } from "../../../services/State";

export enum State {
    WaitTouch = 1,
    SigninInProgress = 2,
    Success = 3,
    Failure = 4,
}

export interface Props {
    authenticationLevel: AuthenticationLevel;

    onRegisterClick: () => void;
    onSignInError: (err: Error) => void;
    onSignInSuccess: (redirectURL: string | undefined) => void;
}

export default function (props: Props) {
    const signInTimeout = 2;
    const [state, setState] = useState(State.WaitTouch);
    const style = useStyles();
    const redirectionURL = useRedirectionURL();
    const mounted = useIsMountedRef();
    const [timerPercent, triggerTimer,] = useTimer(signInTimeout * 1000 - 500);

    const { onSignInSuccess, onSignInError } = props;
    const onSignInErrorCallback = useCallback(onSignInError, []);
    const onSignInSuccessCallback = useCallback(onSignInSuccess, []);

    const doInitiateSignIn = useCallback(async () => {
        // If user is already authenticated, we don't initiate sign in process.
        if (props.authenticationLevel >= AuthenticationLevel.TwoFactor) {
            return;
        }

        try {
            triggerTimer();
            setState(State.WaitTouch);
            const signRequest = await initiateU2FSignin();
            const signRequests: u2fApi.SignRequest[] = [];
            for (var i in signRequest.registeredKeys) {
                const r = signRequest.registeredKeys[i];
                signRequests.push({
                    appId: signRequest.appId,
                    challenge: signRequest.challenge,
                    keyHandle: r.keyHandle,
                    version: r.version,
                })
            }
            const signResponse = await u2fApi.sign(signRequests, signInTimeout);
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;

            setState(State.SigninInProgress);
            const res = await completeU2FSignin(signResponse, redirectionURL);
            setState(State.Success);
            setTimeout(() => { onSignInSuccessCallback(res ? res.redirect : undefined) }, 1500);
        } catch (err) {
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;
            console.error(err);
            onSignInErrorCallback(new Error("Failed to initiate security key sign in process"));
            setState(State.Failure);
        }
    }, [onSignInSuccessCallback, onSignInErrorCallback, redirectionURL, mounted, triggerTimer, props.authenticationLevel]);

    // Set successful state if user is already authenticated.
    useEffect(() => {
        if (props.authenticationLevel >= AuthenticationLevel.TwoFactor) {
            setState(State.Success);
        }
    }, [props.authenticationLevel, setState]);

    useEffect(() => { doInitiateSignIn() }, [doInitiateSignIn]);

    return (
        <MethodContainer
            title="Security Key"
            explanation="Touch the token of your security key"
            onRegisterClick={props.onRegisterClick}>
            <div className={style.icon}>
                <Icon state={state} timer={timerPercent} onRetryClick={doInitiateSignIn} />
            </div>
        </MethodContainer>
    )
}

const useStyles = makeStyles(theme => ({
    icon: {
        display: "inline-block",
    }
}));

interface IconProps {
    state: State;

    timer: number;
    onRetryClick: () => void;
}

function Icon(props: IconProps) {
    const state = props.state as State;
    const theme = useTheme();

    const progressBarStyle: CSSProperties = {
        marginTop: theme.spacing(),
    }

    const touch = <IconWithContext
        icon={<FingerTouchIcon size={64} animated strong />}
        context={<LinearProgressBar value={props.timer} style={progressBarStyle} height={theme.spacing(2)} />}
        className={state === State.WaitTouch ? undefined : "hidden"} />

    const failure = <IconWithContext
        icon={<FailureIcon />}
        context={<Button color="secondary" onClick={props.onRetryClick}>Retry</Button>}
        className={state === State.Failure ? undefined : "hidden"} />

    const success = <IconWithContext
        icon={<SuccessIcon />}
        context={<div style={{ color: "green", padding: theme.spacing() }}>Success!</div>}
        className={state === State.Success || state === State.SigninInProgress ? undefined : "hidden"} />

    return (
        <Fragment>
            {touch}
            {success}
            {failure}
        </Fragment>
    )
}
