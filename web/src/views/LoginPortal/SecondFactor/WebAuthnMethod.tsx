// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useRef } from "react";

import { useTranslation } from "react-i18next";

import WebAuthnTryIcon from "@components/WebAuthnTryIcon";
import { RedirectionURL } from "@constants/SearchParams";
import { useFlow } from "@hooks/Flow";
import { useUserCode } from "@hooks/OpenIDConnect";
import { useQueryParam } from "@hooks/QueryParam";
import { useWebAuthnAssertion } from "@hooks/WebAuthnAssertion";
import { AuthenticationLevel } from "@services/State";
import { getWebAuthnOptions, postWebAuthnResponse } from "@services/WebAuthn";
import MethodContainer, { State as MethodContainerState } from "@views/LoginPortal/SecondFactor/MethodContainer";

export interface Props {
    id: string;
    authenticationLevel: AuthenticationLevel;
    registered: boolean;

    onMethodSelectionClick?: () => void;
    onRegisterClick: () => void;
    onSignInError: (_err: Error) => void;
    onSignInSuccess: (_redirectURL: string | undefined) => void;
}

const WebAuthnMethod = function (props: Props) {
    const { t: translate } = useTranslation();

    const redirectionURL = useQueryParam(RedirectionURL);
    const { flow, id: flowID, subflow } = useFlow();
    const userCode = useUserCode();

    const { ceremony, start, state } = useWebAuthnAssertion({
        getOptions: getWebAuthnOptions,
        onError: props.onSignInError,
        onSuccess: props.onSignInSuccess,
        postResponse: (response, signal) =>
            postWebAuthnResponse(response, redirectionURL, flowID, flow, subflow, userCode, signal),
    });

    const doInitiateSignIn = useCallback(async () => {
        // If user is already authenticated, we don't initiate sign in process.
        if (!props.registered || props.authenticationLevel === AuthenticationLevel.TwoFactor) {
            return;
        }

        await start();
    }, [props.registered, props.authenticationLevel, start]);

    const doInitiateSignInRef = useRef(doInitiateSignIn);

    useEffect(() => {
        doInitiateSignInRef.current = doInitiateSignIn;
    }, [doInitiateSignIn]);

    useEffect(() => {
        doInitiateSignInRef.current().catch(console.error);
    }, []);

    let methodState = MethodContainerState.METHOD;
    if (props.authenticationLevel === AuthenticationLevel.TwoFactor) {
        methodState = MethodContainerState.ALREADY_AUTHENTICATED;
    } else if (!props.registered) {
        methodState = MethodContainerState.NOT_REGISTERED;
    }

    return (
        <MethodContainer
            id={props.id}
            title={translate("Passkey or Security Key")}
            explanation={translate("Use your passkey or security key to continue")}
            duoSelfEnrollment={false}
            registered={props.registered}
            state={methodState}
            onRegisterClick={props.onRegisterClick}
        >
            <WebAuthnTryIcon
                ceremony={ceremony}
                onMethodSelectionClick={props.onMethodSelectionClick}
                onRetryClick={doInitiateSignIn}
                webauthnTouchState={state}
            />
        </MethodContainer>
    );
};

export default WebAuthnMethod;
