import React, { Fragment, useCallback, useEffect, useRef, useState } from "react";

import { Button, makeStyles, useTheme } from "@material-ui/core";
import { CSSProperties } from "@material-ui/styles";
import UAParser from "ua-parser-js";

import FailureIcon from "@components/FailureIcon";
import FingerTouchIcon from "@components/FingerTouchIcon";
import LinearProgressBar from "@components/LinearProgressBar";
import { useIsMountedRef } from "@hooks/Mounted";
import { useRedirectionURL } from "@hooks/RedirectionURL";
import { useTimer } from "@hooks/Timer";
import { AssertionResult } from "@models/Webauthn";
import { AuthenticationLevel } from "@services/State";
import {
    getAssertionPublicKeyCredentialResult,
    getAssertionRequestOptions,
    postAssertionPublicKeyCredentialResult,
} from "@services/Webauthn";
import IconWithContext from "@views/LoginPortal/SecondFactor/IconWithContext";
import MethodContainer, { State as MethodContainerState } from "@views/LoginPortal/SecondFactor/MethodContainer";

export enum State {
    Starting = 1,
    WaitTouch = 2,
    InProgress = 3,
    Failure = 4,
    UserGestureRequired = 5,
}

export interface Props {
    id: string;
    authenticationLevel: AuthenticationLevel;
    registered: boolean;

    onRegisterClick: () => void;
    onSignInError: (err: Error) => void;
    onSignInSuccess: (redirectURL: string | undefined) => void;
}

const WebauthnMethod = function (props: Props) {
    const uaParser = new UAParser();

    const uaResult = uaParser.getResult();

    const apple =
        uaResult.device.vendor === "Apple" ||
        uaResult.os.name === "iOS" ||
        uaResult.os.name === "Mac OS" ||
        uaResult.browser.name === "Safari";

    const signInTimeout = 30;

    const [state, setState] = useState(apple ? State.UserGestureRequired : State.Starting);
    const [explanation, setExplanation] = useState("");

    const style = useStyles();
    const redirectionURL = useRedirectionURL();
    const mounted = useIsMountedRef();
    const [timerPercent, triggerTimer] = useTimer(signInTimeout * 1000 - 500);

    const { onSignInSuccess, onSignInError } = props;
    const onSignInErrorCallback = useRef(onSignInError).current;
    const onSignInSuccessCallback = useRef(onSignInSuccess).current;

    const doInitiateSignIn = useCallback(async () => {
        // If user is already authenticated, we don't initiate sign in process.
        if (!props.registered || props.authenticationLevel === AuthenticationLevel.TwoFactor) {
            return;
        }

        try {
            setState(State.WaitTouch);

            triggerTimer();
            const assertionRequestResponse = await getAssertionRequestOptions();

            if (assertionRequestResponse.status !== 200 || assertionRequestResponse.options == null) {
                setState(State.Failure);
                onSignInErrorCallback(new Error("Failed to initiate security key sign in process"));

                return;
            }

            const result = await getAssertionPublicKeyCredentialResult(assertionRequestResponse.options);

            if (result.result !== AssertionResult.Success) {
                if (!mounted.current) return;
                switch (result.result) {
                    case AssertionResult.FailureUserConsent:
                        onSignInErrorCallback(new Error("You cancelled the assertion request."));
                        break;
                    case AssertionResult.FailureU2FFacetID:
                        onSignInErrorCallback(new Error("The server responded with an invalid Facet ID for the URL."));
                        break;
                    case AssertionResult.FailureSyntax:
                        onSignInErrorCallback(
                            new Error(
                                "The assertion challenge was rejected as malformed or incompatible by your browser.",
                            ),
                        );
                        break;
                    case AssertionResult.FailureWebauthnNotSupported:
                        onSignInErrorCallback(new Error("Your browser does not support the WebAuthN protocol."));
                        break;
                    case AssertionResult.FailureUnknownSecurity:
                        onSignInErrorCallback(new Error("An unknown security error occurred."));
                        break;
                    case AssertionResult.FailureUnknown:
                        onSignInErrorCallback(new Error("An unknown error occurred."));
                        break;
                    default:
                        onSignInErrorCallback(new Error("An unexpected error occurred."));
                        break;
                }
                setState(State.Failure);

                return;
            }

            if (result.credential == null) {
                onSignInErrorCallback(new Error("The browser did not respond with the expected attestation data."));
                setState(State.Failure);

                return;
            }

            if (!mounted.current) return;

            setState(State.InProgress);

            const response = await postAssertionPublicKeyCredentialResult(result.credential, redirectionURL);

            if (response.data.status === "OK" && response.status === 200) {
                onSignInSuccessCallback(response.data.data ? response.data.data.redirect : undefined);
                return;
            }

            if (!mounted.current) return;

            onSignInErrorCallback(new Error("The server rejected the security key."));
            setState(State.Failure);
        } catch (err) {
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;
            console.error(err);
            onSignInErrorCallback(new Error("Failed to initiate security key sign in process"));
            setState(State.Failure);
        }
    }, [
        onSignInErrorCallback,
        onSignInSuccessCallback,
        redirectionURL,
        mounted,
        triggerTimer,
        props.authenticationLevel,
        props.registered,
    ]);

    useEffect(() => {
        if (state === State.Starting) {
            setState(State.WaitTouch);
            doInitiateSignIn();
        }
    }, [doInitiateSignIn, state, setState]);

    useEffect(() => {
        switch (state) {
            case State.UserGestureRequired:
                setExplanation("Click start to initiate the authentication process");
                break;
            case State.Starting:
            case State.WaitTouch:
                setExplanation("Touch the token of your security key");
                break;
            default:
                setExplanation("Something went wrong, click retry to try again");
        }
    }, [state, setExplanation]);

    let methodState = MethodContainerState.METHOD;
    if (props.authenticationLevel === AuthenticationLevel.TwoFactor) {
        methodState = MethodContainerState.ALREADY_AUTHENTICATED;
    } else if (!props.registered) {
        methodState = MethodContainerState.NOT_REGISTERED;
    }

    return (
        <MethodContainer
            id={props.id}
            title="Security Key"
            explanation={explanation}
            duoSelfEnrollment={false}
            registered={props.registered}
            state={methodState}
            onRegisterClick={props.onRegisterClick}
        >
            <div className={style.icon}>
                <Icon state={state} timer={timerPercent} onRetryClick={doInitiateSignIn} />
            </div>
        </MethodContainer>
    );
};

export default WebauthnMethod;

const useStyles = makeStyles(() => ({
    icon: {
        display: "inline-block",
    },
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
    };

    const touch = (
        <IconWithContext
            icon={<FingerTouchIcon size={64} animated strong />}
            className={state === State.WaitTouch || state === State.Starting ? undefined : "hidden"}
        >
            <LinearProgressBar
                value={props.timer}
                style={progressBarStyle}
                className={state === State.WaitTouch ? undefined : "hidden"}
                height={theme.spacing(2)}
            />
        </IconWithContext>
    );

    const waitGesture = (
        <IconWithContext
            icon={<FingerTouchIcon size={64} animated strong />}
            className={state === State.UserGestureRequired ? undefined : "hidden"}
        >
            <Button color="primary" onClick={props.onRetryClick}>
                Start
            </Button>
        </IconWithContext>
    );

    const failure = (
        <IconWithContext icon={<FailureIcon />} className={state === State.Failure ? undefined : "hidden"}>
            <Button color="secondary" onClick={props.onRetryClick}>
                Retry
            </Button>
        </IconWithContext>
    );

    return (
        <Fragment>
            {waitGesture}
            {touch}
            {failure}
        </Fragment>
    );
}
