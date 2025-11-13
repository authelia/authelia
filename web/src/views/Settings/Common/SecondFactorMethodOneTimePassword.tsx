import { Fragment, useCallback, useEffect, useReducer, useRef } from "react";

import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { useUserInfoTOTPConfiguration } from "@hooks/UserInfoTOTPConfiguration";
import { completeTOTPSignIn } from "@services/OneTimePassword";
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

    const timeoutRateLimit = useRef<NodeJS.Timeout | null>(null);

    useEffect(() => {
        if (timeoutRateLimit.current === null) return;

        return clearTimeout(timeoutRateLimit.current);
    }, []);

    useEffect(() => {
        if (fetchConfigError) {
            console.error(fetchConfigError);
            dispatch({ type: "setStatus", status: State.Failure });
        }
    }, [fetchConfigError]);

    useEffect(() => {
        fetchConfig();
    }, [fetchConfig]);

    const handleRateLimited = useCallback(
        (retryAfter: number) => {
            if (timeoutRateLimit.current) {
                clearTimeout(timeoutRateLimit.current);
            }

            dispatch({ type: "setStatus", status: State.RateLimited });

            createErrorNotification(translate("You have made too many requests", { ns: "portal" }));

            timeoutRateLimit.current = setTimeout(() => {
                dispatch({ type: "setStatus", status: State.Idle });
                timeoutRateLimit.current = null;
            }, retryAfter * 1000);
        },
        [createErrorNotification, translate],
    );

    const handleSignIn = useCallback(
        async (passcodeValue: string) => {
            const passcodeStr = `${passcodeValue}`;

            if (!config) return;

            if (!passcodeValue || passcodeStr.length !== config.digits) {
                return;
            }

            try {
                dispatch({ type: "setStatus", status: State.InProgress });

                const res = await completeTOTPSignIn(passcodeStr);

                if (res) {
                    if (res.limited) {
                        handleRateLimited(res.retryAfter);
                    } else {
                        dispatch({ type: "setStatus", status: State.Success });
                        onSecondFactorSuccess();
                    }
                } else {
                    createErrorNotification(translate("The One-Time Password might be wrong", { ns: "portal" }));
                    dispatch({ type: "setStatus", status: State.Failure });
                }
            } catch (err) {
                console.error(err);
                dispatch({ type: "setStatus", status: State.Failure });
            }

            dispatch({ type: "setPasscode", passcode: "" });
        },
        [config, handleRateLimited, createErrorNotification, translate, onSecondFactorSuccess],
    );

    const handlePasscodeChange = useCallback(
        (value: string) => {
            dispatch({ type: "setPasscode", passcode: value });
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
        </Fragment>
    );
};

export default SecondFactorMethodOneTimePassword;
