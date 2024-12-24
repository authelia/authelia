import React, { Fragment, useCallback, useEffect, useState } from "react";

import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { useUserInfoTOTPConfiguration } from "@hooks/UserInfoTOTPConfiguration";
import { completeTOTPSignIn } from "@services/OneTimePassword";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import OTPDial from "@views/LoginPortal/SecondFactor/OTPDial";

export enum State {
    Idle = 1,
    InProgress = 2,
    Success = 3,
    Failure = 4,
}

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

    useEffect(() => {
        if (fetchConfigError) {
            console.error(fetchConfigError);
            setState(State.Failure);
        }
    }, [fetchConfigError, translate]);

    useEffect(() => {
        fetchConfig();
    }, [fetchConfig]);

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
                    createErrorNotification(translate("You have made too many requests"));
                    setState(State.Failure);
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
    }, [passcode, config, createErrorNotification, translate]);

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
