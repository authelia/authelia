import React, { Fragment, useCallback, useEffect, useRef, useState } from "react";

import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { useUserInfoTOTPConfiguration } from "@hooks/UserInfoTOTPConfiguration";
import { completeTOTPSignIn } from "@services/OneTimePassword";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import OTPDial, { State } from "@views/LoginPortal/SecondFactor/OTPDial";

export interface Props {
    closing: boolean;
    onSecondFactorSuccess: () => void;
}

const SecondFactorMethodOneTimePassword = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const [passcode, setPasscode] = useState("");
    const [state, setState] = useState(State.Idle);
    const { createErrorNotification } = useNotifications();

    const [config, fetchConfig, , fetchConfigError] = useUserInfoTOTPConfiguration();

    const timeoutRateLimit = useRef<NodeJS.Timeout>();

    useEffect(() => {
        if (fetchConfigError) {
            console.error(fetchConfigError);
            setState(State.Failure);
        }
    }, [fetchConfigError, translate]);

    useEffect(() => {
        fetchConfig();
    }, [fetchConfig]);

    useEffect(() => {
        return clearTimeout(timeoutRateLimit.current);
    }, []);

    const handleRateLimited = useCallback(
        (retryAfter: number) => {
            if (timeoutRateLimit.current) {
                clearTimeout(timeoutRateLimit.current);
            }

            setState(State.RateLimited);

            createErrorNotification(translate("You have made too many requests"));

            timeoutRateLimit.current = setTimeout(() => {
                setState(State.Idle);
                timeoutRateLimit.current = undefined;
            }, retryAfter * 1000);
        },
        [createErrorNotification, translate],
    );

    const handleSignIn = useCallback(async () => {
        const passcodeStr = `${passcode}`;

        if (!config) return;

        if (!passcode || passcodeStr.length !== config.digits) {
            return;
        }

        try {
            setState(State.InProgress);

            const res = await completeTOTPSignIn(passcodeStr);

            if (res) {
                if (!res.limited) {
                    setState(State.Success);
                } else {
                    handleRateLimited(res.retryAfter);
                }
            } else {
                createErrorNotification(translate("The One-Time Password might be wrong"));
                setState(State.Failure);
            }
        } catch (err) {
            console.error(err);
            setState(State.Failure);
        }

        setPasscode("");
    }, [passcode, config, handleRateLimited, createErrorNotification, translate]);

    useEffect(() => {
        handleSignIn().catch(console.error);
    }, [handleSignIn]);

    return (
        <Fragment>
            {config && !fetchConfigError ? (
                <OTPDial
                    passcode={passcode}
                    period={config.period}
                    digits={config.digits}
                    onChange={setPasscode}
                    state={state}
                />
            ) : (
                <LoadingPage />
            )}
        </Fragment>
    );
};

export default SecondFactorMethodOneTimePassword;
