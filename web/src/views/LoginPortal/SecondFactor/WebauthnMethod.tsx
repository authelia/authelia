import React, { useCallback, useEffect, useRef, useState } from "react";

import WebauthnTryIcon from "@components/WebauthnTryIcon";
import { useIsMountedRef } from "@hooks/Mounted";
import { useRedirectionURL } from "@hooks/RedirectionURL";
import { useWorkflow } from "@hooks/Workflow";
import { AssertionResult, WebauthnTouchState } from "@models/Webauthn";
import { AuthenticationLevel } from "@services/State";
import {
    getAssertionPublicKeyCredentialResult,
    getAssertionRequestOptions,
    postAssertionPublicKeyCredentialResult,
} from "@services/Webauthn";
import MethodContainer, { State as MethodContainerState } from "@views/LoginPortal/SecondFactor/MethodContainer";

export interface Props {
    id: string;
    authenticationLevel: AuthenticationLevel;
    registered: boolean;

    onSignInError: (err: Error) => void;
    onSignInSuccess: (redirectURL: string | undefined) => void;
}

const WebauthnMethod = function (props: Props) {
    const [state, setState] = useState(WebauthnTouchState.WaitTouch);
    const redirectionURL = useRedirectionURL();
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
            const assertionRequestResponse = await getAssertionRequestOptions();

            if (assertionRequestResponse.status !== 200 || assertionRequestResponse.options == null) {
                setState(WebauthnTouchState.Failure);
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
                setState(WebauthnTouchState.Failure);

                return;
            }

            if (result.credential == null) {
                onSignInErrorCallback(new Error("The browser did not respond with the expected attestation data."));
                setState(WebauthnTouchState.Failure);

                return;
            }

            if (!mounted.current) return;

            setState(WebauthnTouchState.InProgress);

            const response = await postAssertionPublicKeyCredentialResult(
                result.credential,
                redirectionURL,
                workflow,
                workflowID,
            );

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
        >
            <WebauthnTryIcon onRetryClick={doInitiateSignIn} webauthnTouchState={state} />
        </MethodContainer>
    );
};

export default WebauthnMethod;
