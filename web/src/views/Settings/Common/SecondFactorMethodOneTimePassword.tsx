// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Fragment, useCallback, useEffect, useReducer } from "react";

import { useTranslation } from "react-i18next";

import RateLimitCountdown from "@components/RateLimitCountdown";
import { useNotifications } from "@contexts/NotificationsContext";
import { useRateLimit } from "@hooks/RateLimit";
import { useUserInfoTOTPConfiguration } from "@hooks/UserInfoTOTPConfiguration";
import { completeTOTPSignIn } from "@services/OneTimePassword";
import { RateLimitedError } from "@services/RateLimit";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import OTPDial, { State } from "@views/LoginPortal/SecondFactor/OTPDial";

type ComponentState = {
    passcode: string;
    status: State;
};

type Action = { type: "setPasscode"; passcode: string } | { type: "setStatus"; status: State };

const initialState: ComponentState = {
    passcode: "",
    status: State.Idle,
};

function reducer(state: ComponentState, action: Action): ComponentState {
    switch (action.type) {
        case "setPasscode":
            return { ...state, passcode: action.passcode };
        case "setStatus":
            return { ...state, status: action.status };
        default:
            return state;
    }
}

export interface Props {
    onSecondFactorSuccess: () => void;
}

const SecondFactorMethodOneTimePassword = function (props: Props) {
    const { onSecondFactorSuccess } = props;
    const { t: translate } = useTranslation(["settings", "portal"]);

    const [state, dispatch] = useReducer(reducer, initialState);
    const { passcode, status } = state;
    const { createErrorNotification } = useNotifications();

    const [config, fetchConfig, , fetchConfigError] = useUserInfoTOTPConfiguration();

    const { secondsRemaining, trigger: triggerRateLimit } = useRateLimit(() =>
        dispatch({ status: State.Idle, type: "setStatus" }),
    );

    useEffect(() => {
        if (fetchConfigError) {
            console.error(fetchConfigError);
            dispatch({ status: State.Failure, type: "setStatus" });
        }
    }, [fetchConfigError]);

    useEffect(() => {
        fetchConfig();
    }, [fetchConfig]);

    const handleRateLimited = useCallback(
        (retryAfter: number) => {
            dispatch({ status: State.RateLimited, type: "setStatus" });

            createErrorNotification(translate("You have made too many requests", { ns: "portal" }));

            triggerRateLimit(retryAfter);
        },
        [createErrorNotification, translate, triggerRateLimit],
    );

    const handleSignIn = useCallback(
        async (passcodeValue: string) => {
            const passcodeStr = `${passcodeValue}`;

            if (!config) return;

            if (!passcodeValue || passcodeStr.length !== config.digits) {
                return;
            }

            try {
                dispatch({ status: State.InProgress, type: "setStatus" });

                await completeTOTPSignIn(passcodeStr);

                dispatch({ status: State.Success, type: "setStatus" });
                onSecondFactorSuccess();
            } catch (err) {
                if (err instanceof RateLimitedError) {
                    handleRateLimited(err.retryAfter);
                } else {
                    console.error(err);
                    dispatch({ status: State.Failure, type: "setStatus" });
                }
            }

            dispatch({ passcode: "", type: "setPasscode" });
        },
        [config, handleRateLimited, onSecondFactorSuccess],
    );

    const handlePasscodeChange = useCallback(
        (value: string) => {
            dispatch({ passcode: value, type: "setPasscode" });
            if (
                config &&
                value.length === config.digits &&
                status !== State.InProgress &&
                status !== State.RateLimited
            ) {
                handleSignIn(value);
            }
        },
        [config, status, handleSignIn],
    );

    return (
        <Fragment>
            {config && !fetchConfigError ? (
                <OTPDial
                    passcode={passcode}
                    period={config.period}
                    digits={config.digits}
                    onChange={handlePasscodeChange}
                    state={status}
                />
            ) : (
                <LoadingPage />
            )}
            <RateLimitCountdown secondsRemaining={secondsRemaining} />
        </Fragment>
    );
};

export default SecondFactorMethodOneTimePassword;
