import { FC, Fragment, KeyboardEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";

import { BroadcastChannel } from "broadcast-channel";
import { Eye, EyeOff } from "lucide-react";
import { useTranslation } from "react-i18next";

import LogoutButton from "@components/LogoutButton";
import SwitchUserButton from "@components/SwitchUserButton";
import { Alert, AlertTitle } from "@components/UI/Alert";
import { Button } from "@components/UI/Button";
import { Input } from "@components/UI/Input";
import { Label } from "@components/UI/Label";
import { Spinner } from "@components/UI/Spinner";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@components/UI/Tooltip";
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
import { cn } from "@utils/Styles";
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
                    <div className="flex flex-col items-center justify-center">
                        <div className="w-full pb-4">
                            <LogoutButton /> {" | "} <SwitchUserButton />
                        </div>
                        <div className="w-full">
                            <div className="flex flex-col items-center justify-center">
                                <div className="w-full">
                                    <div>
                                        <TooltipProvider>
                                            <Tooltip>
                                                <TooltipTrigger asChild>
                                                    <p className="font-semibold">
                                                        {response.client_description === ""
                                                            ? response?.client_id
                                                            : response.client_description}
                                                    </p>
                                                </TooltipTrigger>
                                                <TooltipContent>
                                                    {translate("Client ID", { client_id: response?.client_id }) ||
                                                        "Client ID: " + response?.client_id}
                                                </TooltipContent>
                                            </Tooltip>
                                        </TooltipProvider>
                                    </div>
                                </div>
                                <div className="w-full">
                                    <div>
                                        {translate("The above application is requesting the following permissions")}:
                                    </div>
                                </div>
                                <DecisionFormScopes scopes={response.scopes} />
                                <DecisionFormClaims
                                    claims={claims}
                                    essential_claims={response.essential_claims}
                                    onChangeChecked={(claims) => setClaims(claims)}
                                />
                                {response?.require_login ? (
                                    <div className="my-4 w-full">
                                        <div id={"openid-consent-prompt-login"}>
                                            <div className="grid grid-cols-1 gap-4">
                                                <div className="w-full">
                                                    <TooltipProvider>
                                                        <Tooltip>
                                                            <TooltipTrigger asChild>
                                                                <div>
                                                                    <Label htmlFor="password-textfield">
                                                                        {translate("Password", { ns: "portal" })}
                                                                    </Label>
                                                                    <div className="relative">
                                                                        <Input
                                                                            id={"password-textfield"}
                                                                            ref={passwordRef}
                                                                            onKeyDown={handlePasswordKeyDown}
                                                                            onKeyUp={handlePasswordKeyUp}
                                                                            className={cn(
                                                                                "pr-10",
                                                                                errorPassword && "border-destructive",
                                                                            )}
                                                                            disabled={loading}
                                                                            value={password}
                                                                            onChange={(v) =>
                                                                                setPassword(v.target.value)
                                                                            }
                                                                            onFocus={() => setErrorPassword(false)}
                                                                            type={showPassword ? "text" : "password"}
                                                                            autoComplete={"current-password"}
                                                                            required
                                                                        />
                                                                        <button
                                                                            type="button"
                                                                            className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-muted-foreground hover:text-foreground"
                                                                            aria-label="toggle password visibility"
                                                                            onMouseDown={() => setShowPassword(true)}
                                                                            onMouseUp={() => setShowPassword(false)}
                                                                            onMouseLeave={() => setShowPassword(false)}
                                                                            onTouchStart={() => setShowPassword(true)}
                                                                            onTouchEnd={() => setShowPassword(false)}
                                                                            onTouchCancel={() => setShowPassword(false)}
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
                                                                                <Eye className="h-5 w-5" />
                                                                            ) : (
                                                                                <EyeOff className="h-5 w-5" />
                                                                            )}
                                                                        </button>
                                                                    </div>
                                                                </div>
                                                            </TooltipTrigger>
                                                            <TooltipContent>
                                                                {translate(
                                                                    "You must reauthenticate to be able to give consent",
                                                                )}
                                                            </TooltipContent>
                                                        </Tooltip>
                                                    </TooltipProvider>
                                                </div>
                                                {hasCapsLock ? (
                                                    <div className="mx-2 w-full">
                                                        <Alert variant="default">
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
                                                    </div>
                                                ) : null}
                                            </div>
                                        </div>
                                    </div>
                                ) : null}
                                <OpenIDConnectConsentDecisionFormPreConfiguration
                                    pre_configuration={response.pre_configuration}
                                    onChangePreConfiguration={handlePreConfigureChanged}
                                />
                                <div className="w-full">
                                    <div className="grid grid-cols-2 gap-2">
                                        <div className="w-full">
                                            <TooltipProvider>
                                                <Tooltip>
                                                    <TooltipTrigger asChild>
                                                        <span>
                                                            <Button
                                                                id={"openid-consent-accept"}
                                                                className="mx-2 w-full"
                                                                disabled={!response || passwordMissing || loading}
                                                                onClick={handleAcceptConsent}
                                                                variant={"default"}
                                                            >
                                                                {translate("Accept", { ns: "portal" })}
                                                                {loadingAccept ? (
                                                                    <Spinner className="ml-2 h-5 w-5" />
                                                                ) : null}
                                                            </Button>
                                                        </span>
                                                    </TooltipTrigger>
                                                    <TooltipContent>
                                                        {passwordMissing
                                                            ? translate(
                                                                  "You must reauthenticate to be able to give consent",
                                                              )
                                                            : translate("Accept this consent request")}
                                                    </TooltipContent>
                                                </Tooltip>
                                            </TooltipProvider>
                                        </div>
                                        <div className="w-full">
                                            <TooltipProvider>
                                                <Tooltip>
                                                    <TooltipTrigger asChild>
                                                        <span>
                                                            <Button
                                                                id={"openid-consent-deny"}
                                                                className="mx-2 w-full"
                                                                disabled={!response || loading}
                                                                onClick={handleRejectConsent}
                                                                variant={"secondary"}
                                                            >
                                                                {translate("Deny", { ns: "portal" })}
                                                                {loadingReject ? (
                                                                    <Spinner className="ml-2 h-5 w-5" />
                                                                ) : null}
                                                            </Button>
                                                        </span>
                                                    </TooltipTrigger>
                                                    <TooltipContent>
                                                        {translate("Deny this consent request")}
                                                    </TooltipContent>
                                                </Tooltip>
                                            </TooltipProvider>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </LoginLayout>
            ) : (
                <div>
                    <LoadingPage />
                </div>
            )}
        </Fragment>
    );
};

export default DecisionFormView;
