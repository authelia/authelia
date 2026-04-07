import { useCallback, useEffect, useRef, useState } from "react";

import { Box, Button, CircularProgress, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";

import FailureIcon from "@components/FailureIcon";
import SuccessIcon from "@components/SuccessIcon";
import { RedirectionURL } from "@constants/SearchParams";
import { useFlow } from "@hooks/Flow";
import { useIsMountedRef } from "@hooks/Mounted";
import { useUserCode } from "@hooks/OpenIDConnect";
import { useQueryParam } from "@hooks/QueryParam";
import { AuthenticationLevel } from "@services/State";
import { completeTelegramSignIn, getTelegramStatus, initiateTelegramSignIn } from "@services/Telegram";
import MethodContainer, { State as MethodContainerState } from "@views/LoginPortal/SecondFactor/MethodContainer";

export enum State {
    Idle = 1,
    InProgress = 2,
    Success = 3,
    Failure = 4,
}

export interface Props {
    id: string;
    authenticationLevel: AuthenticationLevel;
    registered: boolean;

    onRegisterClick: () => void;
    onSignInError: (_err: Error) => void;
    onSignInSuccess: (_redirectURL: string | undefined) => void;
}

const TelegramMethod = function (props: Props) {
    const { t: translate } = useTranslation();

    const redirectionURL = useQueryParam(RedirectionURL);
    const { flow, id: flowID, subflow } = useFlow();
    const userCode = useUserCode();
    const mounted = useIsMountedRef();

    const [state, setState] = useState(
        props.authenticationLevel >= AuthenticationLevel.TwoFactor ? State.Success : State.Idle,
    );
    const [, setToken] = useState("");
    const [botDeepLink, setBotDeepLink] = useState("");

    const { onSignInError, onSignInSuccess } = props;
    const onSignInErrorCallback = useRef(onSignInError);
    const onSignInSuccessCallback = useRef(onSignInSuccess);
    const pollIntervalRef = useRef<NodeJS.Timeout | null>(null);

    useEffect(() => {
        onSignInErrorCallback.current = onSignInError;
    }, [onSignInError]);

    useEffect(() => {
        onSignInSuccessCallback.current = onSignInSuccess;
    }, [onSignInSuccess]);

    useEffect(() => {
        return () => {
            if (pollIntervalRef.current !== null) {
                clearInterval(pollIntervalRef.current);
                pollIntervalRef.current = null;
            }
        };
    }, []);

    const derivedState = props.authenticationLevel >= AuthenticationLevel.TwoFactor ? State.Success : state;

    const handleComplete = useCallback(
        async (tkn: string) => {
            try {
                const res = await completeTelegramSignIn(tkn, redirectionURL, flowID, flow, subflow, userCode);
                if (!mounted.current) return;

                setState(State.Success);
                onSignInSuccessCallback.current(res?.redirect);
            } catch (err) {
                if (!mounted.current) return;
                console.error(err);
                onSignInErrorCallback.current(new Error(translate("There was an issue completing sign in process")));
                setState(State.Failure);
            }
        },
        [redirectionURL, flowID, flow, subflow, userCode, mounted, translate],
    );

    const startPolling = useCallback(
        (tkn: string) => {
            if (pollIntervalRef.current !== null) {
                clearInterval(pollIntervalRef.current);
            }

            pollIntervalRef.current = setInterval(async () => {
                try {
                    const status = await getTelegramStatus(tkn);
                    if (!mounted.current) {
                        if (pollIntervalRef.current !== null) {
                            clearInterval(pollIntervalRef.current);
                            pollIntervalRef.current = null;
                        }
                        return;
                    }

                    if (status.expired) {
                        if (pollIntervalRef.current !== null) {
                            clearInterval(pollIntervalRef.current);
                            pollIntervalRef.current = null;
                        }
                        onSignInErrorCallback.current(new Error(translate("The verification has expired")));
                        setState(State.Failure);
                        return;
                    }

                    if (status.verified) {
                        if (pollIntervalRef.current !== null) {
                            clearInterval(pollIntervalRef.current);
                            pollIntervalRef.current = null;
                        }
                        handleComplete(tkn);
                    }
                } catch (err) {
                    console.error(err);
                }
            }, 3000);
        },
        [mounted, translate, handleComplete],
    );

    useEffect(() => {
        if (!props.registered || props.authenticationLevel >= AuthenticationLevel.TwoFactor || state !== State.Idle) {
            return;
        }

        const initiate = async () => {
            try {
                setState(State.InProgress);
                const res = await initiateTelegramSignIn();
                if (!mounted.current) return;

                setToken(res.token);
                setBotDeepLink(res.bot_deep_link);
                startPolling(res.token);
            } catch (err) {
                if (!mounted.current) return;
                console.error(err);
                onSignInErrorCallback.current(
                    new Error(translate("There was an issue initiating Telegram verification")),
                );
                setState(State.Failure);
            }
        };

        initiate().catch(console.error);
    }, [props.registered, props.authenticationLevel, state, mounted, translate, startPolling]);

    const handleRetry = useCallback(() => {
        setToken("");
        setBotDeepLink("");
        setState(State.Idle);
    }, []);

    const effectiveState = derivedState;

    let methodState = MethodContainerState.METHOD;
    if (props.authenticationLevel === AuthenticationLevel.TwoFactor) {
        methodState = MethodContainerState.ALREADY_AUTHENTICATED;
    } else if (!props.registered) {
        methodState = MethodContainerState.NOT_REGISTERED;
    }

    let icon;
    switch (effectiveState) {
        case State.InProgress:
            icon = <CircularProgress size={64} />;
            break;
        case State.Success:
            icon = <SuccessIcon />;
            break;
        case State.Failure:
            icon = <FailureIcon />;
            break;
        default:
            icon = <CircularProgress size={64} />;
    }

    return (
        <MethodContainer
            id={props.id}
            title={translate("Telegram")}
            explanation={translate("Verify your identity via Telegram")}
            duoSelfEnrollment={false}
            registered={props.registered}
            state={methodState}
            onRegisterClick={props.onRegisterClick}
        >
            <Box>{icon}</Box>
            {effectiveState === State.InProgress ? (
                <Box sx={{ marginTop: (theme) => theme.spacing(2) }}>
                    <Typography>{translate("Waiting for verification...")}</Typography>
                    {botDeepLink ? (
                        <Button
                            id={"telegram-open-bot-button"}
                            color="primary"
                            variant="contained"
                            href={botDeepLink}
                            target="_blank"
                            rel="noopener noreferrer"
                            sx={{ marginTop: (theme) => theme.spacing(1) }}
                        >
                            {translate("Open Telegram Bot")}
                        </Button>
                    ) : null}
                </Box>
            ) : null}
            {effectiveState === State.Failure ? (
                <Box>
                    <Button color="secondary" onClick={handleRetry}>
                        {translate("Retry")}
                    </Button>
                </Box>
            ) : null}
        </MethodContainer>
    );
};

export default TelegramMethod;
