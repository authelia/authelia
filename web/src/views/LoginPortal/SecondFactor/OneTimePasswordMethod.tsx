// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useReducer, useRef, useState } from "react";

import { useTranslation } from "react-i18next";

import RateLimitCountdown from "@components/RateLimitCountdown";
import { RedirectionURL } from "@constants/SearchParams";
import { useFlow } from "@hooks/Flow";
import { useUserCode } from "@hooks/OpenIDConnect";
import { useQueryParam } from "@hooks/QueryParam";
import { useRateLimit } from "@hooks/RateLimit";
import { useUserInfoTOTPConfiguration } from "@hooks/UserInfoTOTPConfiguration";
import { completeTOTPSignIn } from "@services/OneTimePassword";
import { RateLimitedError } from "@services/RateLimit";
import { AuthenticationLevel } from "@services/State";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import MethodContainer, { State as MethodContainerState } from "@views/LoginPortal/SecondFactor/MethodContainer";
import OTPDial, { State } from "@views/LoginPortal/SecondFactor/OTPDial";

export interface Props {
    id: string;
    authenticationLevel: AuthenticationLevel;
    registered: boolean;

    onRegisterClick?: () => void;
    onSignInError: (_err: Error) => void;
    onSignInSuccess: (_redirectURL: string | undefined) => void;
}

const OneTimePasswordMethod = function (props: Props) {
    const { t: translate } = useTranslation();

    const redirectionURL = useQueryParam(RedirectionURL);
    const { flow, id: flowID, subflow } = useFlow();
    const userCode = useUserCode();
    const [resp, fetchConfig, , err] = useUserInfoTOTPConfiguration();

    const [passcode, setPasscode] = useState("");

    const stateReducer = useCallback((_state: State, action: { type: State }) => action.type, []);

    const [state, dispatch] = useReducer(
        stateReducer,
        props.authenticationLevel === AuthenticationLevel.TwoFactor ? State.Success : State.Idle,
    );

    const { onSignInError, onSignInSuccess } = props;
    const onSignInErrorCallbackRef = useRef(onSignInError);
    const onSignInSuccessCallbackRef = useRef(onSignInSuccess);
    const { secondsRemaining, trigger: triggerRateLimit } = useRateLimit(() => dispatch({ type: State.Idle }));

    useEffect(() => {
        onSignInErrorCallbackRef.current = onSignInError;
    }, [onSignInError]);

    useEffect(() => {
        onSignInSuccessCallbackRef.current = onSignInSuccess;
    }, [onSignInSuccess]);

    useEffect(() => {
        if (err) {
            console.error(err);
            onSignInErrorCallbackRef.current(new Error(translate("Could not obtain user settings")));
            dispatch({ type: State.Failure });
        }
    }, [onSignInErrorCallbackRef, err, translate]);

    useEffect(() => {
        if (props.registered && props.authenticationLevel === AuthenticationLevel.OneFactor) {
            fetchConfig();
        }
    }, [fetchConfig, props.authenticationLevel, props.registered]);

    const handleRateLimited = useCallback(
        (retryAfter: number) => {
            dispatch({ type: State.RateLimited });

            onSignInErrorCallbackRef.current(new Error(translate("You have made too many requests")));

            triggerRateLimit(retryAfter);
        },
        [onSignInErrorCallbackRef, translate, triggerRateLimit],
    );

    useEffect(() => {
        if (props.authenticationLevel >= AuthenticationLevel.TwoFactor) {
            dispatch({ type: State.Success });
        }
    }, [props.authenticationLevel]);

    useEffect(() => {
        const signInFunc = async () => {
            if (!props.registered || props.authenticationLevel === AuthenticationLevel.TwoFactor) {
                return;
            }

            const passcodeStr = `${passcode}`;

            if (!passcode || passcodeStr.length !== (resp?.digits || 6)) {
                return;
            }

            try {
                dispatch({ type: State.InProgress });
                const res = await completeTOTPSignIn(passcodeStr, redirectionURL, flowID, flow, subflow, userCode);

                dispatch({ type: State.Success });
                onSignInSuccessCallbackRef.current(res?.redirect);
            } catch (err) {
                if (err instanceof RateLimitedError) {
                    handleRateLimited(err.retryAfter);
                } else {
                    console.error(err);
                    onSignInErrorCallbackRef.current(new Error(translate("The One-Time Password might be wrong")));
                    dispatch({ type: State.Failure });
                }
            }
            setPasscode("");
        };

        signInFunc().catch(console.error);
    }, [
        props.registered,
        props.authenticationLevel,
        passcode,
        resp?.digits,
        redirectionURL,
        flowID,
        flow,
        subflow,
        userCode,
        onSignInErrorCallbackRef,
        translate,
        onSignInSuccessCallbackRef,
        handleRateLimited,
    ]);

    let methodState = MethodContainerState.METHOD;
    if (props.authenticationLevel === AuthenticationLevel.TwoFactor) {
        methodState = MethodContainerState.ALREADY_AUTHENTICATED;
    } else if (!props.registered) {
        methodState = MethodContainerState.NOT_REGISTERED;
    }

    return (
        <MethodContainer
            id={props.id}
            title={translate("One-Time Password")}
            explanation={translate("Enter One-Time Password")}
            duoSelfEnrollment={false}
            registered={props.registered}
            state={methodState}
            onRegisterClick={props.onRegisterClick}
        >
            <div>
                {resp !== undefined || err !== undefined ? (
                    <OTPDial
                        passcode={passcode}
                        period={resp?.period || 30}
                        digits={resp?.digits || 6}
                        onChange={setPasscode}
                        state={state}
                    />
                ) : (
                    <LoadingPage />
                )}
                <RateLimitCountdown secondsRemaining={secondsRemaining} />
            </div>
        </MethodContainer>
    );
};

export default OneTimePasswordMethod;
