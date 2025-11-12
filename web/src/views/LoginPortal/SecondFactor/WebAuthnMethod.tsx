import { useCallback, useEffect, useReducer, useRef } from "react";

import { useTranslation } from "react-i18next";

import WebAuthnTryIcon from "@components/WebAuthnTryIcon";
import { RedirectionURL } from "@constants/SearchParams";
import { useFlow } from "@hooks/Flow";
import { useIsMountedRef } from "@hooks/Mounted";
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
    onSignInError: (err: Error) => void;
    onSignInSuccess: (redirectURL: string | undefined) => void;
}

const WebAuthnMethod = function (props: Props) {
    const { t: translate } = useTranslation();

    const redirectionURL = useQueryParam(RedirectionURL);
    const { flow, id: flowID, subflow } = useFlow();
    const userCode = useUserCode();
    const mounted = useIsMountedRef();

    const reducer = (state: WebAuthnTouchState, action: { type: "setState"; state: WebAuthnTouchState }) => {
        if (action.type === "setState") {
            return action.state;
        }
        return state;
    };

    const [state, dispatch] = useReducer(reducer, WebAuthnTouchState.WaitTouch);

    const { onSignInError, onSignInSuccess } = props;
    const onSignInErrorCallback = useRef(onSignInError).current;
    const onSignInSuccessCallback = useRef(onSignInSuccess).current;

    const doInitiateSignIn = useCallback(async () => {
        // If user is already authenticated, we don't initiate sign in process.
        if (!props.registered || props.authenticationLevel === AuthenticationLevel.TwoFactor) {
            return;
        }

        try {
            dispatch({ state: WebAuthnTouchState.WaitTouch, type: "setState" });
            const optionsStatus = await getWebAuthnOptions();

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                dispatch({ state: WebAuthnTouchState.Failure, type: "setState" });
                onSignInErrorCallback(new Error(translate("Failed to initiate security key sign in process")));

                return;
            }

            const result = await getWebAuthnResult(optionsStatus.options);

            if (result.result !== AssertionResult.Success) {
                if (!mounted.current) return;

                dispatch({ state: WebAuthnTouchState.Failure, type: "setState" });

                onSignInErrorCallback(new Error(translate(AssertionResultFailureString(result.result))));

                return;
            }

            if (result.response == null) {
                onSignInErrorCallback(
                    new Error(translate("The browser did not respond with the expected attestation data")),
                );
                dispatch({ state: WebAuthnTouchState.Failure, type: "setState" });

                return;
            }

            if (!mounted.current) return;

            dispatch({ state: WebAuthnTouchState.InProgress, type: "setState" });

            const response = await postWebAuthnResponse(
                result.response,
                redirectionURL,
                flowID,
                flow,
                subflow,
                userCode,
            );

            if (response.data.status === "OK" && response.status === 200) {
                onSignInSuccessCallback(response.data.data ? response.data.data.redirect : undefined);
                return;
            }

            if (!mounted.current) return;

            onSignInErrorCallback(new Error(translate("The server rejected the security key")));
            dispatch({ state: WebAuthnTouchState.Failure, type: "setState" });
        } catch (err) {
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;
            console.error(err);
            onSignInErrorCallback(new Error(translate("Failed to initiate security key sign in process")));
            dispatch({ state: WebAuthnTouchState.Failure, type: "setState" });
        }
    }, [
        props.registered,
        props.authenticationLevel,
        mounted,
        redirectionURL,
        flowID,
        flow,
        subflow,
        userCode,
        onSignInErrorCallback,
        translate,
        onSignInSuccessCallback,
    ]);

    useEffect(() => {
        doInitiateSignIn().catch(console.error);
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
