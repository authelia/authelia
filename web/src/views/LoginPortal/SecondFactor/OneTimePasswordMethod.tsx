import React, { useCallback, useEffect, useRef, useState } from "react";

import { Box } from "@mui/material";
import { useTranslation } from "react-i18next";

import { RedirectionURL } from "@constants/SearchParams";
import { useFlow } from "@hooks/Flow";
import { useQueryParam } from "@hooks/QueryParam";
import { useUserInfoTOTPConfiguration } from "@hooks/UserInfoTOTPConfiguration";
import { completeTOTPSignIn } from "@services/OneTimePassword";
import { AuthenticationLevel } from "@services/State";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import MethodContainer, { State as MethodContainerState } from "@views/LoginPortal/SecondFactor/MethodContainer";
import OTPDial, { State } from "@views/LoginPortal/SecondFactor/OTPDial";

export interface Props {
    id: string;
    authenticationLevel: AuthenticationLevel;
    registered: boolean;

    onRegisterClick: () => void;
    onSignInError: (err: Error) => void;
    onSignInSuccess: (redirectURL: string | undefined) => void;
}

const OneTimePasswordMethod = function (props: Props) {
    const [passcode, setPasscode] = useState("");
    const [state, setState] = useState(
        props.authenticationLevel === AuthenticationLevel.TwoFactor ? State.Success : State.Idle,
    );
    const redirectionURL = useQueryParam(RedirectionURL);
    const { id: flowID, flow, subflow } = useFlow();
    const { t: translate } = useTranslation();

    const { onSignInSuccess, onSignInError } = props;
    const onSignInErrorCallback = useRef(onSignInError).current;
    const onSignInSuccessCallback = useRef(onSignInSuccess).current;
    const [resp, fetch, , err] = useUserInfoTOTPConfiguration();

    const timeoutRateLimit = useRef<NodeJS.Timeout | null>(null);

    useEffect(() => {
        if (timeoutRateLimit.current === null) return;

        return clearTimeout(timeoutRateLimit.current);
    }, []);

    useEffect(() => {
        if (err) {
            console.error(err);
            onSignInErrorCallback(new Error(translate("Could not obtain user settings")));
            setState(State.Failure);
        }
    }, [onSignInErrorCallback, err, translate]);

    useEffect(() => {
        if (props.registered && props.authenticationLevel === AuthenticationLevel.OneFactor) {
            fetch();
        }
    }, [fetch, props.authenticationLevel, props.registered]);

    const handleRateLimited = useCallback(
        (retryAfter: number) => {
            if (timeoutRateLimit.current) {
                clearTimeout(timeoutRateLimit.current);
            }

            setState(State.RateLimited);

            onSignInErrorCallback(new Error(translate("You have made too many requests")));

            timeoutRateLimit.current = setTimeout(() => {
                setState(State.Idle);
                timeoutRateLimit.current = null;
            }, retryAfter * 1000);
        },
        [onSignInErrorCallback, translate],
    );

    const signInFunc = useCallback(async () => {
        if (!props.registered || props.authenticationLevel === AuthenticationLevel.TwoFactor) {
            return;
        }

        const passcodeStr = `${passcode}`;

        if (!passcode || passcodeStr.length !== (resp?.digits || 6)) {
            return;
        }

        try {
            setState(State.InProgress);
            const res = await completeTOTPSignIn(passcodeStr, redirectionURL, flowID, flow, subflow);

            if (!res) {
                onSignInErrorCallback(new Error(translate("The One-Time Password might be wrong")));
                setState(State.Failure);
            } else if (!res.limited) {
                setState(State.Success);
                onSignInSuccessCallback(res && res.data ? res.data.redirect : undefined);
            } else {
                handleRateLimited(res.retryAfter);
            }
        } catch (err) {
            console.error(err);
            onSignInErrorCallback(new Error(translate("The One-Time Password might be wrong")));
            setState(State.Failure);
        }
        setPasscode("");
    }, [
        props.registered,
        props.authenticationLevel,
        passcode,
        resp?.digits,
        redirectionURL,
        flowID,
        flow,
        subflow,
        onSignInErrorCallback,
        translate,
        onSignInSuccessCallback,
        handleRateLimited,
    ]);

    // Set successful state if user is already authenticated.
    useEffect(() => {
        if (props.authenticationLevel >= AuthenticationLevel.TwoFactor) {
            setState(State.Success);
        }
    }, [props.authenticationLevel, setState]);

    useEffect(() => {
        signInFunc().catch(console.error);
    }, [signInFunc]);

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
            <Box>
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
            </Box>
        </MethodContainer>
    );
};

export default OneTimePasswordMethod;
