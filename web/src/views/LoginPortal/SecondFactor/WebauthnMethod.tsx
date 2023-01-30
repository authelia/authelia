import React, { useCallback, useEffect, useRef, useState } from "react";

import WebauthnTryIcon from "@components/WebauthnTryIcon";
import { RedirectionURL } from "@constants/SearchParams";
import { useIsMountedRef } from "@hooks/Mounted";
import { useQueryParam } from "@hooks/QueryParam";
import { useWorkflow } from "@hooks/Workflow";
import { AssertionResult, AssertionResultFailureString, WebauthnTouchState } from "@models/Webauthn";
import { AuthenticationLevel } from "@services/State";
import { getAuthenticationOptions, getAuthenticationResult, postAuthenticationResponse } from "@services/Webauthn";
import MethodContainer, { State as MethodContainerState } from "@views/LoginPortal/SecondFactor/MethodContainer";

export interface Props {
    id: string;
    authenticationLevel: AuthenticationLevel;
    registered: boolean;

    onRegisterClick: () => void;
    onSignInError: (err: Error) => void;
    onSignInSuccess: (redirectURL: string | undefined) => void;
}

const WebauthnMethod = function (props: Props) {
    const [state, setState] = useState(WebauthnTouchState.WaitTouch);
    const redirectionURL = useQueryParam(RedirectionURL);
    const [workflow, workflowID] = useWorkflow();
    const mounted = useIsMountedRef();

    const { onSignInSuccess, onSignInError } = props;
    const onSignInErrorCallback = useRef(onSignInError).current;
    const onSignInSuccessCallback = useRef(onSignInSuccess).current;

    const doInitiateSignIn = useCallback(async () => {
        // If user is already authenticated, we don't initiate sign in process.
        if (!props.registered || props.authenticationLevel === AuthenticationLevel.TwoFactor) {
            return;
        }

        try {
            setState(WebauthnTouchState.WaitTouch);
            const optionsStatus = await getAuthenticationOptions();

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                setState(WebauthnTouchState.Failure);
                onSignInErrorCallback(new Error("Failed to initiate security key sign in process"));

                return;
            }

            const result = await getAuthenticationResult(optionsStatus.options);

            if (result.result !== AssertionResult.Success) {
                if (!mounted.current) return;

                setState(WebauthnTouchState.Failure);

                onSignInErrorCallback(new Error(AssertionResultFailureString(result.result)));

                return;
            }

            if (result.response == null) {
                onSignInErrorCallback(new Error("The browser did not respond with the expected attestation data."));
                setState(WebauthnTouchState.Failure);

                return;
            }

            if (!mounted.current) return;

            setState(WebauthnTouchState.InProgress);

            const response = await postAuthenticationResponse(result.response, redirectionURL, workflow, workflowID);

            if (response.data.status === "OK" && response.status === 200) {
                onSignInSuccessCallback(response.data.data ? response.data.data.redirect : undefined);
                return;
            }

            if (!mounted.current) return;

            onSignInErrorCallback(new Error("The server rejected the security key."));
            setState(WebauthnTouchState.Failure);
        } catch (err) {
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;
            console.error(err);
            onSignInErrorCallback(new Error("Failed to initiate security key sign in process"));
            setState(WebauthnTouchState.Failure);
        }
    }, [
        onSignInErrorCallback,
        onSignInSuccessCallback,
        redirectionURL,
        workflow,
        workflowID,
        mounted,
        props.authenticationLevel,
        props.registered,
    ]);

    useEffect(() => {
        doInitiateSignIn();
    }, [doInitiateSignIn]);

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
            explanation="Touch the token of your security key"
            duoSelfEnrollment={false}
            registered={props.registered}
            state={methodState}
            onRegisterClick={props.onRegisterClick}
        >
            <WebauthnTryIcon onRetryClick={doInitiateSignIn} webauthnTouchState={state} />
        </MethodContainer>
    );
};

export default WebauthnMethod;
