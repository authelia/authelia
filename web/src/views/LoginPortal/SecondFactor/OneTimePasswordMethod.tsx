import { useCallback, useEffect, useReducer, useRef, useState } from "react";

import { Box } from "@mui/material";
import { useTranslation } from "react-i18next";

import { RedirectionURL } from "@constants/SearchParams";
import { useFlow } from "@hooks/Flow";
import { useUserCode } from "@hooks/OpenIDConnect";
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
    onSignInError: (_err: Error) => void;
    onSignInSuccess: (_redirectURL: string | undefined) => void;
}

const OneTimePasswordMethod = function (props: Props) {
    const { t: translate } = useTranslation();

    const redirectionURL = useQueryParam(RedirectionURL);
    const { flow, id: flowID, subflow } = useFlow();
    const userCode = useUserCode();
    const [resp, fetch, , err] = useUserInfoTOTPConfiguration();

    const [passcode, setPasscode] = useState("");

    const stateReducer = useCallback((_state: State, action: { type: State }) => action.type, []);

    const [state, dispatch] = useReducer(
        stateReducer,
        props.authenticationLevel === AuthenticationLevel.TwoFactor ? State.Success : State.Idle,
    );

    const { onSignInError, onSignInSuccess } = props;
    const onSignInErrorCallback = useRef(onSignInError);
    const onSignInSuccessCallback = useRef(onSignInSuccess);
    const timeoutRateLimit = useRef<NodeJS.Timeout | null>(null);

    useEffect(() => {
        onSignInErrorCallback.current = onSignInError;
    }, [onSignInError]);

    useEffect(() => {
        onSignInSuccessCallback.current = onSignInSuccess;
    }, [onSignInSuccess]);

    useEffect(() => {
        return () => {
            if (timeoutRateLimit.current !== null) {
                clearTimeout(timeoutRateLimit.current);
                timeoutRateLimit.current = null;
            }
        };
    }, []);

    useEffect(() => {
        if (err) {
            console.error(err);
            onSignInErrorCallback.current(new Error(translate("Could not obtain user settings")));
            dispatch({ type: State.Failure });
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

            dispatch({ type: State.RateLimited });

            onSignInErrorCallback.current(new Error(translate("You have made too many requests")));

            timeoutRateLimit.current = setTimeout(() => {
                dispatch({ type: State.Idle });
                timeoutRateLimit.current = null;
            }, retryAfter * 1000);
        },
        [onSignInErrorCallback, translate],
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

                if (!res) {
                    onSignInErrorCallback.current(new Error(translate("The One-Time Password might be wrong")));
                    dispatch({ type: State.Failure });
                } else if (res.limited) {
                    handleRateLimited(res.retryAfter);
                } else {
                    dispatch({ type: State.Success });
                    onSignInSuccessCallback.current(res?.data?.redirect);
                }
            } catch (err) {
                console.error(err);
                onSignInErrorCallback.current(new Error(translate("The One-Time Password might be wrong")));
                dispatch({ type: State.Failure });
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
        onSignInErrorCallback,
        translate,
        onSignInSuccessCallback,
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
