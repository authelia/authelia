import { useCallback, useEffect, useRef, useState } from "react";

import { useFormStatus } from "react-dom";
import { useTranslation } from "react-i18next";

import LogoutButton from "@components/LogoutButton";
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
import { useUserCode } from "@hooks/OpenIDConnect";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import LoginLayout from "@layouts/LoginLayout";
import { AutheliaState, AuthenticationLevel } from "@services/State";
import LoadingPage from "@views/LoadingPage/LoadingPage";

const normalizeUserCode = (value: string) => value.toUpperCase().replace(/\s+/g, "");

export interface Props {
    state: AutheliaState;
}

function DeviceAuthorizationFormView({ state }: Props) {
    const { t: translate } = useTranslation(["consent", "settings"]);

    const userCode = useUserCode();

    const [code, setCode] = useState(() => (userCode ? normalizeUserCode(userCode) : ""));

    const navigate = useRouterNavigate();

    const autoSubmittedRef = useRef(false);

    const handleCode = useCallback(
        (code: string) => {
            if (code === "") {
                return;
            }

            const params = new URLSearchParams();

            params.set(UserCode, code);
            params.set(Flow, FlowNameOpenIDConnect);
            params.set(SubFlow, SubFlowNameDeviceAuthorization);

            navigate(`${ConsentRoute}${ConsentOpenIDSubRoute}${ConsentDecisionSubRoute}`, true, true, true, params);
        },
        [navigate],
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

    const submitCode = useCallback(() => {
        handleCode(code);
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
                                maxLength={UserCodeLength}
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
                        <DeviceAuthorizationSubmit disabled={code === ""}>
                            {translate("Confirm", { ns: "settings" })}
                        </DeviceAuthorizationSubmit>
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
