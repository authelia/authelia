import { FC, useCallback, useEffect, useRef, useState } from "react";

import { useTranslation } from "react-i18next";

import LogoutButton from "@components/LogoutButton";
import SwitchUserButton from "@components/SwitchUserButton";
import { Button } from "@components/UI/Button";
import { Input } from "@components/UI/Input";
import { Label } from "@components/UI/Label";
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

export interface Props {
    state: AutheliaState;
}

const DeviceAuthorizationFormView: FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation(["consent", "settings"]);

    const userCode = useUserCode();

    const [code, setCode] = useState(userCode || "");

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
        if (props.state.authentication_level === AuthenticationLevel.Unauthenticated) {
            const params = new URLSearchParams();

            if (userCode) {
                params.set(UserCode, userCode);
            }

            params.set(Flow, FlowNameOpenIDConnect);
            params.set(SubFlow, SubFlowNameDeviceAuthorization);

            navigate(IndexRoute, true, true, true, params);
        }
    }, [userCode, navigate, props.state.authentication_level]);

    useEffect(() => {
        autoSubmittedRef.current = false;
    }, [userCode]);

    useEffect(() => {
        if (
            !userCode ||
            props.state.authentication_level === AuthenticationLevel.Unauthenticated ||
            autoSubmittedRef.current
        ) {
            return;
        }

        autoSubmittedRef.current = true;
        handleCode(userCode);
    }, [handleCode, props.state.authentication_level, userCode]);

    return props.state.authentication_level === AuthenticationLevel.Unauthenticated ? (
        <div>
            <LoadingPage />
        </div>
    ) : (
        <LoginLayout id={"openid-consent-device-auth-stage"} title={translate("Confirm the Code")}>
            <div className="flex flex-col items-center justify-center">
                <div className="w-full pb-4">
                    <LogoutButton /> {" | "} <SwitchUserButton />
                </div>
                <div className="w-full">
                    <div id={"form-consent-openid-device-code-authorization"}>
                        <div className="grid grid-cols-1 gap-4">
                            <div className="w-full">
                                <Label htmlFor="user-code">{translate("Code")}</Label>
                                <Input
                                    id={"user-code"}
                                    required
                                    value={code}
                                    onChange={(v) => setCode(v.target.value)}
                                    autoCapitalize={"none"}
                                />
                            </div>
                            <div className="w-full">
                                <Button
                                    id={"confirm-button"}
                                    variant={"default"}
                                    className="w-full"
                                    onClick={() => handleCode(code)}
                                    disabled={code === ""}
                                >
                                    {translate("Confirm", { ns: "settings" })}
                                </Button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </LoginLayout>
    );
};

export default DeviceAuthorizationFormView;
