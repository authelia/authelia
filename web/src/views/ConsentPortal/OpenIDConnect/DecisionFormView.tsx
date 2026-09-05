import { ReactNode, useActionState, useCallback, useEffect, useMemo, useRef, useState } from "react";

import { BroadcastChannel } from "broadcast-channel";
import { useFormStatus } from "react-dom";
import { useTranslation } from "react-i18next";

import LogoutButton from "@components/LogoutButton";
import SwitchUserButton from "@components/SwitchUserButton";
import { Button, ButtonColor } from "@components/UI/Button";
import { Card } from "@components/UI/Card";
import { Separator } from "@components/UI/Separator";
import { Spinner } from "@components/UI/Spinner";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@components/UI/Tooltip";
import { ConsentCompletionSubRoute, ConsentRoute, IndexRoute } from "@constants/Routes";
import {
    Decision,
    DecisionAccepted,
    DecisionRejected,
    Flow,
    SubFlow,
    SubFlowNameDeviceAuthorization,
} from "@constants/SearchParams";
import { useNotifications } from "@contexts/NotificationsContext";
import { useFlow } from "@hooks/Flow";
import { useUserCode } from "@hooks/OpenIDConnect";
import { useRedirector } from "@hooks/Redirector";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import LoginLayout from "@layouts/LoginLayout";
import { UserInfo } from "@models/UserInfo";
import {
    ConsentGetResponseBody,
    getConsentResponse,
    postConsentResponseAccept,
    postConsentResponseReject,
    putDeviceCodeFlowUserCode,
} from "@services/ConsentOpenIDConnect";
import { postFirstFactorReauthenticate } from "@services/Password";
import { AutheliaState, AuthenticationLevel } from "@services/State";
import DecisionFormPreConfiguration from "@views/ConsentPortal/OpenIDConnect/DecisionFormPreConfiguration";
import DecisionFormReauthentication, {
    Props as ReauthenticationProps,
} from "@views/ConsentPortal/OpenIDConnect/DecisionFormReauthentication";
import DecisionFormRequest from "@views/ConsentPortal/OpenIDConnect/DecisionFormRequest";
import LoadingPage from "@views/LoadingPage/LoadingPage";

const FieldDecision = "decision";

const DecisionAccept = "accept";
const DecisionDeny = "deny";
const DecisionAuthenticate = "authenticate";
const DecisionCancel = "cancel";

type Step = "authenticate" | "decision";

export interface Props {
    state: AutheliaState;
    userInfo?: UserInfo;
}

function DecisionFormView({ state, userInfo }: Props) {
    const { t: translate } = useTranslation(["consent", "portal", "settings"]);

    const { createErrorNotification } = useNotifications();
    const navigate = useRouterNavigate();
    const redirect = useRedirector();
    const { flow, id: flowID, subflow } = useFlow();
    const userCode = useUserCode();

    const [response, setResponse] = useState<ConsentGetResponseBody>();
    const [error, setError] = useState<any>(undefined);
    const [claims, setClaims] = useState<string[]>([]);
    const [preConfigure, setPreConfigure] = useState(false);
    const [step, setStep] = useState<Step>("decision");
    const [password, setPassword] = useState("");
    const [errorPassword, setErrorPassword] = useState(false);
    const [failure, setFailure] = useState<null | string>(null);

    const loginChannel = useMemo(() => new BroadcastChannel<boolean>("login"), []);

    const passwordRef = useRef<HTMLInputElement | null>(null);

    useEffect(() => {
        if (state.authentication_level === AuthenticationLevel.Unauthenticated) {
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
    }, [flowID, navigate, state.authentication_level, userCode]);

    useEffect(() => {
        if (error) {
            navigate(IndexRoute);
            console.error(`Unable to display consent screen: ${error.message}`);
        }
    }, [navigate, error]);

    const focusPassword = useCallback(() => {
        if (passwordRef.current === null) return;

        passwordRef.current.focus();
    }, [passwordRef]);

    useEffect(() => {
        if (step !== "authenticate") return;

        const timeout = setTimeout(() => focusPassword(), 10);

        return () => clearTimeout(timeout);
    }, [focusPassword, step]);

    const navigateCompletion = useCallback(
        (decision: string) => {
            const query = new URLSearchParams();

            if (flow) {
                query.set(Flow, flow);
            }

            if (subflow) {
                query.set(SubFlow, subflow);
            }

            query.set(Decision, decision);

            navigate(ConsentRoute + ConsentCompletionSubRoute, false, false, false, query);
        },
        [flow, navigate, subflow],
    );

    const submitAcceptance = useCallback(async () => {
        if (!response) {
            return;
        }

        const res = await postConsentResponseAccept(
            preConfigure,
            response.client_id,
            claims,
            flowID,
            subflow,
            userCode,
        );

        if (subflow === SubFlowNameDeviceAuthorization) {
            if (res.flow_id && userCode) {
                await putDeviceCodeFlowUserCode(res.flow_id, userCode);

                navigateCompletion(DecisionAccepted);
            } else {
                createErrorNotification(translate("Failed to submit the user code"));
            }
        } else if (res.redirect_uri) {
            redirect(res.redirect_uri);
        } else {
            createErrorNotification(translate("Failed to redirect you", { ns: "portal" }));
        }
    }, [
        claims,
        createErrorNotification,
        flowID,
        navigateCompletion,
        preConfigure,
        redirect,
        response,
        subflow,
        translate,
        userCode,
    ]);

    const handleAccept = useCallback(async () => {
        if (!response) {
            return;
        }

        if (response.require_login) {
            setStep("authenticate");

            return;
        }

        await submitAcceptance();
    }, [response, submitAcceptance]);

    const handleAuthenticate = useCallback(async () => {
        if (!response) {
            return;
        }

        if (password.length === 0) {
            setErrorPassword(true);

            focusPassword();

            return;
        }

        const fail = (message: string) => {
            setFailure(message);
            setPassword("");
            focusPassword();
        };

        try {
            await postFirstFactorReauthenticate(password, undefined, undefined, flowID, flow, subflow, userCode);
            await loginChannel.postMessage(true);
        } catch (err) {
            console.error(err);

            fail(translate("Incorrect password", { ns: "portal" }));

            return;
        }

        const r = await getConsentResponse(flowID, userCode);

        setResponse(r);

        if (r.require_login) {
            fail(translate("Failed to confirm your identity", { ns: "portal" }));

            return;
        }

        await submitAcceptance();
    }, [flow, flowID, focusPassword, loginChannel, password, response, subflow, submitAcceptance, translate, userCode]);

    const handleCancel = useCallback(() => {
        setStep("decision");
        setPassword("");
        setErrorPassword(false);
        setFailure(null);
    }, []);

    const handleReject = useCallback(async () => {
        if (!response) {
            return;
        }

        const res = await postConsentResponseReject(response.client_id, flowID, subflow, userCode);

        if (subflow === SubFlowNameDeviceAuthorization) {
            navigateCompletion(DecisionRejected);
        } else if (res.redirect_uri) {
            redirect(res.redirect_uri);
        } else {
            createErrorNotification(translate("Failed to redirect you", { ns: "portal" }));
        }
    }, [createErrorNotification, flowID, navigateCompletion, redirect, response, subflow, translate, userCode]);

    const handlePasswordChange = useCallback((value: string) => {
        setErrorPassword(false);
        setFailure(null);
        setPassword(value);
    }, []);

    const [, submitDecision] = useActionState(async (_: null, data: FormData) => {
        switch (data.get(FieldDecision)) {
            case DecisionDeny:
                await handleReject();

                break;
            case DecisionAuthenticate:
                await handleAuthenticate();

                break;
            case DecisionCancel:
                handleCancel();

                break;
            default:
                await handleAccept();

                break;
        }

        return null;
    }, null);

    if (!userInfo || response === undefined) {
        return (
            <div>
                <LoadingPage />
            </div>
        );
    }

    const authenticating = step === "authenticate";

    return (
        <LoginLayout
            id={"openid-consent-decision-stage"}
            title={`${translate("Hi", { ns: "portal" })} ${userInfo.display_name}`}
            subtitle={translate("Consent Request")}
            maxWidth={"sm"}
        >
            <div className="flex w-full flex-col gap-4">
                <div className="flex w-full items-center justify-center">
                    <LogoutButton />
                    <div className="flex h-4 items-center">
                        <Separator orientation={"vertical"} />
                    </div>
                    <SwitchUserButton />
                </div>
                <form action={submitDecision} className="flex w-full flex-col gap-6">
                    <DecisionFormRequest
                        response={response}
                        claims={claims}
                        onChangeClaims={setClaims}
                        collapsible={authenticating}
                    />
                    <div className="flex w-full flex-col gap-3">
                        {authenticating ? (
                            <Card className="gap-0 px-4 py-4">
                                <p className="mb-2 text-left text-sm text-muted-foreground">
                                    {translate("Enter your password to confirm your identity", { ns: "portal" })}
                                </p>
                                <DecisionFormReauthenticationField
                                    ref={passwordRef}
                                    value={password}
                                    error={errorPassword}
                                    failure={failure}
                                    onChange={handlePasswordChange}
                                />
                            </Card>
                        ) : (
                            <DecisionFormPreConfiguration
                                pre_configuration={response.pre_configuration}
                                checked={preConfigure}
                                onChangePreConfiguration={setPreConfigure}
                            />
                        )}
                        <div className="grid grid-cols-2 gap-2">
                            {authenticating ? (
                                <>
                                    <DecisionFormButton
                                        id={"openid-consent-authenticate"}
                                        decision={DecisionAuthenticate}
                                        color={"primary"}
                                        tooltip={translate("You must reauthenticate to be able to give consent")}
                                    >
                                        {translate("Submit", { ns: "settings" })}
                                    </DecisionFormButton>
                                    <DecisionFormButton
                                        id={"openid-consent-cancel"}
                                        decision={DecisionCancel}
                                        color={"secondary"}
                                        variant={"outline"}
                                        formNoValidate
                                        tooltip={translate("Return to the consent request")}
                                    >
                                        {translate("Cancel", { ns: "portal" })}
                                    </DecisionFormButton>
                                </>
                            ) : (
                                <>
                                    <DecisionFormButton
                                        id={"openid-consent-accept"}
                                        decision={DecisionAccept}
                                        color={"primary"}
                                        tooltip={translate("Accept this consent request")}
                                    >
                                        {translate("Accept", { ns: "portal" })}
                                    </DecisionFormButton>
                                    <DecisionFormButton
                                        id={"openid-consent-deny"}
                                        decision={DecisionDeny}
                                        color={"secondary"}
                                        variant={"outline"}
                                        formNoValidate
                                        tooltip={translate("Deny this consent request")}
                                    >
                                        {translate("Deny", { ns: "portal" })}
                                    </DecisionFormButton>
                                </>
                            )}
                        </div>
                    </div>
                </form>
            </div>
        </LoginLayout>
    );
}

function DecisionFormReauthenticationField(props: Omit<ReauthenticationProps, "disabled">) {
    const { pending } = useFormStatus();

    return <DecisionFormReauthentication {...props} disabled={pending} />;
}

interface DecisionFormButtonProps {
    children: ReactNode;
    color: ButtonColor;
    variant?: "default" | "outline";
    decision: string;
    disabled?: boolean;
    formNoValidate?: boolean;
    id: string;
    tooltip: string;
}

function DecisionFormButton({
    children,
    color,
    decision,
    disabled,
    formNoValidate,
    id,
    tooltip,
    variant = "default",
}: DecisionFormButtonProps) {
    const { data, pending } = useFormStatus();

    const active = pending && data?.get(FieldDecision) === decision;

    return (
        <TooltipProvider>
            <Tooltip>
                <TooltipTrigger
                    render={
                        <span>
                            <Button
                                id={id}
                                type={"submit"}
                                name={FieldDecision}
                                value={decision}
                                formNoValidate={formNoValidate}
                                className="w-full"
                                disabled={disabled || pending}
                                variant={variant}
                                color={color}
                            >
                                {children}
                                {active ? <Spinner data-testid={"spinner"} size={20} className="ml-2 h-5 w-5" /> : null}
                            </Button>
                        </span>
                    }
                />
                <TooltipContent sideOffset={8}>{tooltip}</TooltipContent>
            </Tooltip>
        </TooltipProvider>
    );
}

export default DecisionFormView;
