import React, { Fragment, useCallback, useEffect, useRef, useState } from "react";

import { useTranslation } from "react-i18next";

import { useIsMountedRef } from "@hooks/Mounted";
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
    const { t: translate } = useTranslation(["settings", "portal"]);

    const [passcode, setPasscode] = useState("");
    const [state, setState] = useState(State.Idle);
    const { createErrorNotification } = useNotifications();

    const [config, fetchConfig, , fetchConfigError] = useUserInfoTOTPConfiguration();
    const mounted = useIsMountedRef();

    const timeoutRateLimit = useRef<NodeJS.Timeout | null>(null);

    useEffect(() => {
        return () => {
            if (timeoutRateLimit.current !== null) {
                clearTimeout(timeoutRateLimit.current);
                timeoutRateLimit.current = null;
            }
        };
    }, []);

    useEffect(() => {
        if (fetchConfigError) {
            console.error(fetchConfigError);
            setState(State.Failure);
        }
    }, [fetchConfigError, translate]);

    useEffect(() => {
        fetchConfig();
    }, [fetchConfig]);

    const handleRateLimited = useCallback(
        (retryAfter: number) => {
            if (timeoutRateLimit.current) {
                clearTimeout(timeoutRateLimit.current);
            }

            setState(State.RateLimited);

            createErrorNotification(translate("You have made too many requests", { ns: "portal" }));

            timeoutRateLimit.current = setTimeout(() => {
                if (!mounted.current) {
                    timeoutRateLimit.current = null;
                    return;
                }
                setState(State.Idle);
                timeoutRateLimit.current = null;
            }, retryAfter * 1000);
        },
        [createErrorNotification, translate, mounted],
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

            if (!mounted.current) {
                return;
            }

            if (res) {
                if (!res.limited) {
                    setState(State.Success);
                } else {
                    handleRateLimited(res.retryAfter);
                }
            } else {
                createErrorNotification(translate("The One-Time Password might be wrong", { ns: "portal" }));
                setState(State.Failure);
            }
        } catch (err) {
            console.error(err);
            if (!mounted.current) {
                return;
            }
            setState(State.Failure);
        }

        setPasscode("");
    }, [passcode, config, handleRateLimited, createErrorNotification, translate, mounted]);

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
