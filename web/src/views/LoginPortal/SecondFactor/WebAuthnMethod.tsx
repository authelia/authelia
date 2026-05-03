import { useCallback, useEffect, useReducer, useRef } from "react";

import axios from "axios";
import { useTranslation } from "react-i18next";

import WebAuthnTryIcon from "@components/WebAuthnTryIcon";
import { RedirectionURL } from "@constants/SearchParams";
import { useAbortSignal } from "@hooks/Abort";
import { useFlow } from "@hooks/Flow";
import { useUserCode } from "@hooks/OpenIDConnect";
import { useQueryParam } from "@hooks/QueryParam";
import { AssertionResult, AssertionResultFailureString, WebAuthnTouchState } from "@models/WebAuthn";
import { AuthenticationLevel } from "@services/State";
import { getWebAuthnOptions, getWebAuthnResult, postWebAuthnResponse } from "@services/WebAuthn";
import MethodContainer, { State as MethodContainerState } from "@views/LoginPortal/SecondFactor/MethodContainer";

export interface Props {
    id: string;
    authenticationLevel: AuthenticationLevel;
    registered: boolean;

    onRegisterClick: () => void;
    onSignInError: (_err: Error) => void;
    onSignInSuccess: (_redirectURL: string | undefined) => void;
}

const WebAuthnMethod = function (props: Props) {
    const { t: translate } = useTranslation();

    const redirectionURL = useQueryParam(RedirectionURL);
    const { flow, id: flowID, subflow } = useFlow();
    const userCode = useUserCode();
    const getSignal = useAbortSignal();

    const stateReducer = (_state: WebAuthnTouchState, action: { type: WebAuthnTouchState }) => action.type;

    const [state, dispatch] = useReducer(stateReducer, WebAuthnTouchState.WaitTouch);

    const { onSignInError, onSignInSuccess } = props;
    const signInInitiatedRef = useRef(false);

    const doInitiateSignIn = useCallback(async () => {
        // If user is already authenticated, we don't initiate sign in process.
        if (!props.registered || props.authenticationLevel === AuthenticationLevel.TwoFactor) {
            return;
        }

        const signal = getSignal();

        try {
            dispatch({ type: WebAuthnTouchState.WaitTouch });
            const optionsStatus = await getWebAuthnOptions(signal);

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                dispatch({ type: WebAuthnTouchState.Failure });
                onSignInError(new Error(translate("Failed to initiate security key sign in process")));

                return;
            }

            const result = await getWebAuthnResult(optionsStatus.options);

            if (signal.aborted) return;

            if (result.result !== AssertionResult.Success) {
                dispatch({ type: WebAuthnTouchState.Failure });

                onSignInError(new Error(translate(AssertionResultFailureString(result.result))));

                return;
            }

            if (result.response == null) {
                onSignInError(new Error(translate("The browser did not respond with the expected attestation data")));
                dispatch({ type: WebAuthnTouchState.Failure });

                return;
            }

            dispatch({ type: WebAuthnTouchState.InProgress });

            const response = await postWebAuthnResponse(
                result.response,
                redirectionURL,
                flowID,
                flow,
                subflow,
                userCode,
                signal,
            );

            if (response.data.status === "OK" && response.status === 200) {
                onSignInSuccess(response.data.data ? response.data.data.redirect : undefined);
                return;
            }

            onSignInError(new Error(translate("The server rejected the security key")));
            dispatch({ type: WebAuthnTouchState.Failure });
        } catch (err) {
            if (axios.isCancel(err)) return;
            console.error(err);
            onSignInError(new Error(translate("Failed to initiate security key sign in process")));
            dispatch({ type: WebAuthnTouchState.Failure });
        }
    }, [
        props.registered,
        props.authenticationLevel,
        getSignal,
        redirectionURL,
        flowID,
        flow,
        subflow,
        userCode,
        onSignInError,
        translate,
        onSignInSuccess,
    ]);

    useEffect(() => {
        if (!signInInitiatedRef.current) {
            signInInitiatedRef.current = true;
            doInitiateSignIn().catch(console.error);
        }
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
            title={translate("Security Key")}
            explanation={translate("Touch the token of your security key")}
            duoSelfEnrollment={false}
            registered={props.registered}
            state={methodState}
            onRegisterClick={props.onRegisterClick}
        >
            <WebAuthnTryIcon onRetryClick={doInitiateSignIn} webauthnTouchState={state} />
        </MethodContainer>
    );
};

export default WebAuthnMethod;
