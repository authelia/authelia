import { FC, Fragment, KeyboardEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";

import { Visibility, VisibilityOff } from "@mui/icons-material";
import {
    Alert,
    AlertTitle,
    Box,
    Button,
    CircularProgress,
    FormControl,
    IconButton,
    InputAdornment,
    Theme,
    Tooltip,
    Typography,
    useTheme,
} from "@mui/material";
import Grid from "@mui/material/Grid";
import TextField from "@mui/material/TextField";
import { BroadcastChannel } from "broadcast-channel";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

import LogoutButton from "@components/LogoutButton";
import SwitchUserButton from "@components/SwitchUserButton";
import { ConsentCompletionSubRoute, ConsentRoute, IndexRoute } from "@constants/Routes";
import { Decision, Flow, SubFlow, SubFlowNameDeviceAuthorization } from "@constants/SearchParams";
import { useFlow } from "@hooks/Flow";
import { useNotifications } from "@hooks/NotificationsContext";
import { useUserCode } from "@hooks/OpenIDConnect";
import { useRedirector } from "@hooks/Redirector";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import LoginLayout from "@layouts/LoginLayout";
import { UserInfo } from "@models/UserInfo";
import { IsCapsLockModified } from "@services/CapsLock";
import {
    ConsentGetResponseBody,
    getConsentResponse,
    postConsentResponseAccept,
    postConsentResponseReject,
    putDeviceCodeFlowUserCode,
} from "@services/ConsentOpenIDConnect";
import { postFirstFactorReauthenticate } from "@services/Password";
import { AutheliaState, AuthenticationLevel } from "@services/State";
import DecisionFormClaims from "@views/ConsentPortal/OpenIDConnect/DecisionFormClaims";
import OpenIDConnectConsentDecisionFormPreConfiguration from "@views/ConsentPortal/OpenIDConnect/DecisionFormPreConfiguration";
import DecisionFormScopes from "@views/ConsentPortal/OpenIDConnect/DecisionFormScopes";
import LoadingPage from "@views/LoadingPage/LoadingPage";

export interface Props {
    userInfo?: UserInfo;
    state: AutheliaState;
}

const DecisionFormView: FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation(["consent", "portal"]);
    const theme = useTheme();

    const { classes } = useStyles();

    const { createErrorNotification, resetNotification } = useNotifications();
    const navigate = useRouterNavigate();
    const redirect = useRedirector();
    const { flow, id: flowID, subflow } = useFlow();
    const userCode = useUserCode();

    const [password, setPassword] = useState("");
    const [hasCapsLock, setHasCapsLock] = useState(false);
    const [isCapsLockPartial, setIsCapsLockPartial] = useState(false);
    const [loading, setLoading] = useState(false);
    const [loadingAccept, setLoadingAccept] = useState(false);
    const [loadingReject, setLoadingReject] = useState(false);
    const [errorPassword, setErrorPassword] = useState(false);
    const [showPassword, setShowPassword] = useState(false);

    const [response, setResponse] = useState<ConsentGetResponseBody>();
    const [error, setError] = useState<any>(undefined);
    const [claims, setClaims] = useState<string[]>([]);
    const [preConfigure, setPreConfigure] = useState(false);

    const loginChannel = useMemo(() => new BroadcastChannel<boolean>("login"), []);

    const passwordRef = useRef<HTMLInputElement | null>(null);

    const handlePreConfigureChanged = (value: boolean) => {
        setPreConfigure(value);
    };

    useEffect(() => {
        if (props.state.authentication_level === AuthenticationLevel.Unauthenticated) {
            navigate(IndexRoute);
        } else if (flowID || userCode) {
            getConsentResponse(flowID, userCode)
                .then((r) => {
                    setResponse(r);
                    setClaims(r.claims || []);
                })
                .catch((error) => {
                    setError(error);
                });
        } else {
            navigate(IndexRoute);
        }
    }, [flowID, navigate, props.state.authentication_level, userCode]);

    useEffect(() => {
        if (error) {
            navigate(IndexRoute);
            console.error(`Unable to display consent screen: ${error.message}`);
        }
    }, [navigate, resetNotification, createErrorNotification, error]);

    const focusPassword = useCallback(() => {
        if (passwordRef.current === null) return;

        passwordRef.current.focus();
    }, [passwordRef]);

    const handleAcceptConsent = useCallback(async () => {
        // This case should not happen in theory because the buttons are disabled when response is undefined.
        if (!response) {
            return;
        }

        if (response.require_login) {
            if (password.length === 0) {
                setErrorPassword(true);

                focusPassword();

                return;
            }

            setLoading(true);
            setLoadingAccept(true);

            try {
                await postFirstFactorReauthenticate(password, undefined, undefined, flowID, flow, subflow, userCode);
                await loginChannel.postMessage(true);
            } catch (err) {
                console.error(err);
                createErrorNotification(translate("Failed to confirm your identity", { ns: "portal" }));
                setPassword("");
                setLoading(false);
                setLoadingAccept(false);
                focusPassword();

                return;
            }

            const r = await getConsentResponse(flowID, userCode);

            setResponse(r);

            if (r.require_login) {
                createErrorNotification(translate("Failed to confirm your identity", { ns: "portal" }));

                return;
            }
        } else {
            setLoading(true);
            setLoadingAccept(true);
        }

        const res = await postConsentResponseAccept(
            preConfigure,
            response.client_id,
            claims,
            flowID,
            subflow,
            userCode,
        );

        setLoading(false);
        setLoadingAccept(false);

        if ((!subflow || subflow === "") && res.redirect_uri) {
            redirect(res.redirect_uri);
        } else if (subflow && subflow === SubFlowNameDeviceAuthorization) {
            if (res.flow_id && userCode) {
                await putDeviceCodeFlowUserCode(res.flow_id, userCode);

                const query = new URLSearchParams();

                if (flow) {
                    query.set(Flow, flow);
                }

                if (subflow) {
                    query.set(SubFlow, subflow);
                }

                query.set(Decision, "accepted");

                navigate(ConsentRoute + ConsentCompletionSubRoute, false, false, false, query);
            } else {
                createErrorNotification(translate("Failed to submit the user code"));
                throw new Error("Failed to perform user code submission");
            }
        } else {
            createErrorNotification(translate("Failed to redirect you", { ns: "portal" }));
            throw new Error("Unable to redirect the user");
        }
    }, [
        claims,
        createErrorNotification,
        flow,
        flowID,
        focusPassword,
        loginChannel,
        navigate,
        password,
        preConfigure,
        redirect,
        response,
        subflow,
        translate,
        userCode,
    ]);

    const handleRejectConsent = async () => {
        if (!response) {
            return;
        }

        setLoading(true);
        setLoadingReject(true);

        const res = await postConsentResponseReject(response.client_id, flowID, subflow, userCode);

        setLoading(false);
        setLoadingReject(false);

        if ((!subflow || subflow === "") && res.redirect_uri) {
            redirect(res.redirect_uri);
        } else if (subflow && subflow === SubFlowNameDeviceAuthorization) {
            const query = new URLSearchParams();

            if (flow) {
                query.set(Flow, flow);
            }

            if (subflow) {
                query.set(SubFlow, subflow);
            }

            query.set(Decision, "rejected");

            navigate(ConsentRoute + ConsentCompletionSubRoute, false, false, false, query);
        } else {
            throw new Error("Unable to redirect the user");
        }
    };

    useEffect(() => {
        const timeout = setTimeout(() => focusPassword(), 10);
        return () => clearTimeout(timeout);
    }, [focusPassword]);

    const handlePasswordKeyDown = useCallback(
        (event: KeyboardEvent<HTMLDivElement>) => {
            if (event.key === "Enter") {
                event.preventDefault();

                if (password.length === 0) {
                    focusPassword();
                } else {
                    handleAcceptConsent().catch(console.error);
                }
            }
        },
        [focusPassword, handleAcceptConsent, password.length],
    );

    const handlePasswordKeyUp = useCallback(
        (event: KeyboardEvent<HTMLDivElement>) => {
            if (password.length <= 1) {
                setHasCapsLock(false);
                setIsCapsLockPartial(false);

                if (password.length === 0) {
                    return;
                }
            }

            const modified = IsCapsLockModified(event);

            if (modified === null) return;

            if (modified) {
                setHasCapsLock(true);
            } else {
                setIsCapsLockPartial(true);
            }
        },
        [password.length],
    );

    const passwordMissing = response?.require_login && password.length === 0;

    return (
        <Fragment>
            {props.userInfo && response !== undefined ? (
                <LoginLayout
                    id={"openid-consent-decision-stage"}
                    title={`${translate("Hi", { ns: "portal" })} ${props.userInfo.display_name}`}
                    subtitle={translate("Consent Request")}
                >
                    <Grid container direction={"column"} justifyContent={"center"} alignItems={"center"}>
                        <Grid size={{ xs: 12 }} sx={{ paddingBottom: theme.spacing(2) }}>
                            <LogoutButton /> {" | "} <SwitchUserButton />
                        </Grid>
                        <Grid size={{ xs: 12 }}>
                            <Grid container alignItems={"center"} justifyContent={"center"}>
                                <Grid size={{ xs: 12 }}>
                                    <Box>
                                        <Tooltip
                                            title={
                                                translate("Client ID", { client_id: response?.client_id }) ||
                                                "Client ID: " + response?.client_id
                                            }
                                        >
                                            <Typography className={classes.clientDescription}>
                                                {response.client_description === ""
                                                    ? response?.client_id
                                                    : response.client_description}
                                            </Typography>
                                        </Tooltip>
                                    </Box>
                                </Grid>
                                <Grid size={{ xs: 12 }}>
                                    <Box>
                                        {translate("The above application is requesting the following permissions")}:
                                    </Box>
                                </Grid>
                                <DecisionFormScopes scopes={response.scopes} scopeDescriptions={response.scope_descriptions} />
                                <DecisionFormClaims
                                    claims={claims}
                                    essential_claims={response.essential_claims}
                                    onChangeChecked={(claims) => setClaims(claims)}
                                />
                                {response?.require_login ? (
                                    <Grid size={{ xs: 12 }} marginY={theme.spacing(2)}>
                                        <FormControl id={"openid-consent-prompt-login"}>
                                            <Grid container spacing={2}>
                                                <Grid size={{ xs: 12 }}>
                                                    <Tooltip
                                                        title={translate(
                                                            "You must reauthenticate to be able to give consent",
                                                        )}
                                                    >
                                                        <TextField
                                                            id={"password-textfield"}
                                                            label={translate("Password", { ns: "portal" })}
                                                            variant={"outlined"}
                                                            inputRef={passwordRef}
                                                            onKeyDown={handlePasswordKeyDown}
                                                            onKeyUp={handlePasswordKeyUp}
                                                            error={errorPassword}
                                                            disabled={loading}
                                                            value={password}
                                                            onChange={(v) => setPassword(v.target.value)}
                                                            onFocus={() => setErrorPassword(false)}
                                                            type={showPassword ? "text" : "password"}
                                                            autoComplete={"current-password"}
                                                            required
                                                            fullWidth
                                                            slotProps={{
                                                                input: {
                                                                    endAdornment: (
                                                                        <InputAdornment position="end">
                                                                            <IconButton
                                                                                aria-label="toggle password visibility"
                                                                                edge="end"
                                                                                size="large"
                                                                                onMouseDown={() =>
                                                                                    setShowPassword(true)
                                                                                }
                                                                                onMouseUp={() => setShowPassword(false)}
                                                                                onMouseLeave={() =>
                                                                                    setShowPassword(false)
                                                                                }
                                                                                onTouchStart={() =>
                                                                                    setShowPassword(true)
                                                                                }
                                                                                onTouchEnd={() =>
                                                                                    setShowPassword(false)
                                                                                }
                                                                                onTouchCancel={() =>
                                                                                    setShowPassword(false)
                                                                                }
                                                                                onKeyDown={(e) => {
                                                                                    if (e.key === " ") {
                                                                                        setShowPassword(true);
                                                                                        e.preventDefault();
                                                                                    }
                                                                                }}
                                                                                onKeyUp={(e) => {
                                                                                    if (e.key === " ") {
                                                                                        setShowPassword(false);
                                                                                        e.preventDefault();
                                                                                    }
                                                                                }}
                                                                            >
                                                                                {showPassword ? (
                                                                                    <Visibility />
                                                                                ) : (
                                                                                    <VisibilityOff />
                                                                                )}
                                                                            </IconButton>
                                                                        </InputAdornment>
                                                                    ),
                                                                },
                                                            }}
                                                        />
                                                    </Tooltip>
                                                </Grid>
                                                {hasCapsLock ? (
                                                    <Grid size={{ xs: 12 }} marginX={2}>
                                                        <Alert severity={"warning"}>
                                                            <AlertTitle>
                                                                {translate("Warning", { ns: "portal" })}
                                                            </AlertTitle>
                                                            {isCapsLockPartial
                                                                ? translate(
                                                                      "The password was partially entered with Caps Lock",
                                                                      { ns: "portal" },
                                                                  )
                                                                : translate("The password was entered with Caps Lock", {
                                                                      ns: "portal",
                                                                  })}
                                                        </Alert>
                                                    </Grid>
                                                ) : null}
                                            </Grid>
                                        </FormControl>
                                    </Grid>
                                ) : null}
                                <OpenIDConnectConsentDecisionFormPreConfiguration
                                    pre_configuration={response.pre_configuration}
                                    onChangePreConfiguration={handlePreConfigureChanged}
                                />
                                <Grid size={{ xs: 12 }}>
                                    <Grid container spacing={1}>
                                        <Grid size={{ xs: 6 }}>
                                            <Tooltip
                                                title={
                                                    passwordMissing
                                                        ? translate(
                                                              "You must reauthenticate to be able to give consent",
                                                          )
                                                        : translate("Accept this consent request")
                                                }
                                            >
                                                <span>
                                                    <Button
                                                        id={"openid-consent-accept"}
                                                        className={classes.button}
                                                        disabled={!response || passwordMissing || loading}
                                                        onClick={handleAcceptConsent}
                                                        color={"primary"}
                                                        variant={"contained"}
                                                        endIcon={loadingAccept ? <CircularProgress size={20} /> : null}
                                                    >
                                                        {translate("Accept", { ns: "portal" })}
                                                    </Button>
                                                </span>
                                            </Tooltip>
                                        </Grid>
                                        <Grid size={{ xs: 6 }}>
                                            <Tooltip title={translate("Deny this consent request")}>
                                                <span>
                                                    <Button
                                                        id={"openid-consent-deny"}
                                                        className={classes.button}
                                                        disabled={!response || loading}
                                                        onClick={handleRejectConsent}
                                                        color={"secondary"}
                                                        variant={"contained"}
                                                        endIcon={loadingReject ? <CircularProgress size={20} /> : null}
                                                    >
                                                        {translate("Deny", { ns: "portal" })}
                                                    </Button>
                                                </span>
                                            </Tooltip>
                                        </Grid>
                                    </Grid>
                                </Grid>
                            </Grid>
                        </Grid>
                    </Grid>
                </LoginLayout>
            ) : (
                <Box>
                    <LoadingPage />
                </Box>
            )}
        </Fragment>
    );
};

const useStyles = makeStyles()((theme: Theme) => ({
    button: {
        marginLeft: theme.spacing(),
        marginRight: theme.spacing(),
        width: "100%",
    },
    clientDescription: {
        fontWeight: 600,
    },
}));

export default DecisionFormView;
