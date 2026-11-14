// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useMemo, useRef } from "react";

import { useTranslation } from "react-i18next";

import { RedirectionURL } from "@constants/SearchParams";
import { useFlow } from "@hooks/Flow";
import { useUserCode } from "@hooks/OpenIDConnect";
import { useQueryParam } from "@hooks/QueryParam";
import { AppAuthenticatorSignInBody } from "@services/AppAuthenticator";
import { AuthenticationLevel } from "@services/State";
import AppAuthenticatorChallenge from "@views/LoginPortal/SecondFactor/AppAuthenticatorChallenge";
import MethodContainer, { State as MethodContainerState } from "@views/LoginPortal/SecondFactor/MethodContainer";

export interface Props {
    id: string;
    authenticationLevel: AuthenticationLevel;
    registered: boolean;

    onRegisterClick: () => void;
    onSignInError: (_err: Error) => void;
    onSignInSuccess: (_redirectURL: string | undefined) => void;
}

const AppAuthenticatorMethod = function (props: Props) {
    const { t: translate } = useTranslation();

    const { flow, id: flowID, subflow } = useFlow();
    const userCode = useUserCode();
    const redirectionURL = useQueryParam(RedirectionURL);

    const { onSignInSuccess } = props;
    const timeoutSuccessRef = useRef<null | ReturnType<typeof setTimeout>>(null);

    useEffect(() => {
        return () => {
            if (timeoutSuccessRef.current !== null) {
                clearTimeout(timeoutSuccessRef.current);
                timeoutSuccessRef.current = null;
            }
        };
    }, []);

    const body: AppAuthenticatorSignInBody = useMemo(
        () => ({ flow, flowID, subflow, targetURL: redirectionURL, userCode }),
        [flow, flowID, subflow, redirectionURL, userCode],
    );

    const handleSuccess = useCallback(
        (redirect: string | undefined) => {
            timeoutSuccessRef.current = setTimeout(() => {
                timeoutSuccessRef.current = null;
                onSignInSuccess(redirect);
            }, 1500);
        },
        [onSignInSuccess],
    );

    const methodState =
        props.authenticationLevel >= AuthenticationLevel.TwoFactor
            ? MethodContainerState.ALREADY_AUTHENTICATED
            : MethodContainerState.METHOD;

    return (
        <MethodContainer
            id={props.id}
            title={translate("Authelia app")}
            explanation={""}
            duoSelfEnrollment={false}
            registered={props.registered}
            state={methodState}
            onRegisterClick={props.onRegisterClick}
        >
            {methodState !== MethodContainerState.METHOD ? null : props.registered ? (
                <AppAuthenticatorChallenge body={body} onSuccess={handleSuccess} onError={props.onSignInError} />
            ) : (
                <p id="app-authenticator-not-registered">{translate("Register a device in settings")}</p>
            )}
        </MethodContainer>
    );
};

export default AppAuthenticatorMethod;
