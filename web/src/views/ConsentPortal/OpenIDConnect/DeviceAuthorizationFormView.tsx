// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useRef, useState } from "react";

import { useFormStatus } from "react-dom";
import { useTranslation } from "react-i18next";

import LogoutButton from "@components/LogoutButton";
import RateLimitCountdown from "@components/RateLimitCountdown";
import SwitchUserButton from "@components/SwitchUserButton";
import { Button } from "@components/UI/Button";
import { Card } from "@components/UI/Card";
import { Field, FieldDescription, FieldLabel } from "@components/UI/Field";
import { Input } from "@components/UI/Input";
import { Separator } from "@components/UI/Separator";
import { Spinner } from "@components/UI/Spinner";
import { UserCodeLength } from "@constants/OpenIDConnect";
import { ConsentDecisionSubRoute, ConsentOpenIDSubRoute, ConsentRoute, IndexRoute } from "@constants/Routes";
import {
    Flow,
    FlowNameOpenIDConnect,
    SubFlow,
    SubFlowNameDeviceAuthorization,
    UserCode,
} from "@constants/SearchParams";
import { useNotifications } from "@contexts/NotificationsContext";
import { useUserCode } from "@hooks/OpenIDConnect";
import { useRateLimit } from "@hooks/RateLimit";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import LoginLayout from "@layouts/LoginLayout";
import { getConsentResponse } from "@services/ConsentOpenIDConnect";
import { RateLimitedError } from "@services/RateLimit";
import { AutheliaState, AuthenticationLevel } from "@services/State";
import LoadingPage from "@views/LoadingPage/LoadingPage";

const normalizeUserCode = (value: string) => value.toUpperCase().replace(/\s+/g, "").slice(0, UserCodeLength);

export interface Props {
    state: AutheliaState;
}

function DeviceAuthorizationFormView({ state }: Props) {
    const { t: translate } = useTranslation(["consent", "settings"]);

    const userCode = useUserCode();

    const [code, setCode] = useState(() => (userCode ? normalizeUserCode(userCode) : ""));

    const navigate = useRouterNavigate();

    const { createErrorNotification } = useNotifications();

    const { limited: rateLimited, secondsRemaining, trigger: triggerRateLimit } = useRateLimit();

    const autoSubmittedRef = useRef(false);

    const handleCode = useCallback(
        async (code: string) => {
            if (code === "") {
                return;
            }

            try {
                await getConsentResponse(undefined, code);
            } catch (err) {
                if (err instanceof RateLimitedError) {
                    triggerRateLimit(err.retryAfter);

                    createErrorNotification(translate("You have made too many requests", { ns: "portal" }));
                } else {
                    createErrorNotification(translate("The code is invalid or has expired"));
                }

                return;
            }

            const params = new URLSearchParams();

            params.set(UserCode, code);
            params.set(Flow, FlowNameOpenIDConnect);
            params.set(SubFlow, SubFlowNameDeviceAuthorization);

            navigate(`${ConsentRoute}${ConsentOpenIDSubRoute}${ConsentDecisionSubRoute}`, true, true, true, params);
        },
        [createErrorNotification, navigate, translate, triggerRateLimit],
    );

    useEffect(() => {
        if (state.authentication_level === AuthenticationLevel.Unauthenticated) {
            const params = new URLSearchParams();

            if (userCode) {
                params.set(UserCode, userCode);
            }

            params.set(Flow, FlowNameOpenIDConnect);
            params.set(SubFlow, SubFlowNameDeviceAuthorization);

            navigate(IndexRoute, true, true, true, params);
        }
    }, [userCode, navigate, state.authentication_level]);

    useEffect(() => {
        autoSubmittedRef.current = false;
    }, [userCode]);

    useEffect(() => {
        if (
            !userCode ||
            state.authentication_level === AuthenticationLevel.Unauthenticated ||
            autoSubmittedRef.current
        ) {
            return;
        }

        autoSubmittedRef.current = true;
        handleCode(normalizeUserCode(userCode));
    }, [handleCode, state.authentication_level, userCode]);

    const submitCode = useCallback(async () => {
        await handleCode(code);
    }, [code, handleCode]);

    if (state.authentication_level === AuthenticationLevel.Unauthenticated) {
        return (
            <div>
                <LoadingPage />
            </div>
        );
    }

    return (
        <LoginLayout id={"openid-consent-device-auth-stage"} title={translate("Confirm the Code")}>
            <div className="flex w-full flex-col gap-4">
                <div className="flex w-full items-center justify-center">
                    <LogoutButton />
                    <div className="flex h-4 items-center">
                        <Separator orientation={"vertical"} />
                    </div>
                    <SwitchUserButton />
                </div>
                <form
                    id={"form-consent-openid-device-code-authorization"}
                    action={submitCode}
                    className="flex w-full flex-col gap-6"
                >
                    <Card className="gap-0 px-4 py-4">
                        <Field className="text-left">
                            <FieldLabel htmlFor="user-code">{translate("Code")}</FieldLabel>
                            <FieldDescription>{translate("Enter the code displayed on your device")}</FieldDescription>
                            <Input
                                id={"user-code"}
                                name={"user_code"}
                                value={code}
                                onChange={(event) => setCode(normalizeUserCode(event.target.value))}
                                className="text-center indent-[0.2em] font-mono text-lg tracking-[0.2em] uppercase"
                                autoCapitalize={"characters"}
                                autoComplete={"one-time-code"}
                                spellCheck={false}
                                aria-required={"true"}
                            />
                        </Field>
                    </Card>
                    <div className="flex w-full flex-col gap-3">
                        <p className="text-xs text-muted-foreground">
                            {translate("You will be asked to review the request next")}
                        </p>
                        <DeviceAuthorizationSubmit disabled={code === "" || rateLimited}>
                            {translate("Confirm", { ns: "settings" })}
                        </DeviceAuthorizationSubmit>
                        <RateLimitCountdown secondsRemaining={secondsRemaining} />
                    </div>
                </form>
            </div>
        </LoginLayout>
    );
}

function DeviceAuthorizationSubmit({ children, disabled }: { children: string; disabled: boolean }) {
    const { pending } = useFormStatus();

    return (
        <Button
            id={"confirm-button"}
            type={"submit"}
            variant={"default"}
            color={"primary"}
            className="w-full"
            disabled={disabled || pending}
        >
            {children}
            {pending ? <Spinner data-testid={"spinner"} size={20} className="ml-2 h-5 w-5" /> : null}
        </Button>
    );
}

export default DeviceAuthorizationFormView;
