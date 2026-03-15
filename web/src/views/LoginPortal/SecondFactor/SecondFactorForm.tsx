import { lazy, useState } from "react";

import { browserSupportsWebAuthn } from "@simplewebauthn/browser";
import { useTranslation } from "react-i18next";
import { Route, Routes } from "react-router-dom";

import LogoutButton from "@components/LogoutButton";
import SwitchUserButton from "@components/SwitchUserButton";
import { Button } from "@components/UI/Button";
import {
    SecondFactorPasswordSubRoute,
    SecondFactorPushSubRoute,
    SecondFactorTOTPSubRoute,
    SecondFactorWebAuthnSubRoute,
    SettingsRoute,
    SettingsTwoFactorAuthenticationSubRoute,
} from "@constants/Routes";
import { useLocalStorageMethodContext } from "@contexts/LocalStorageMethodContext";
import { useFlowPresent } from "@hooks/Flow";
import { useNotifications } from "@hooks/NotificationsContext";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import LoginLayout from "@layouts/LoginLayout";
import { Configuration } from "@models/Configuration";
import { SecondFactorMethod } from "@models/Methods";
import { UserInfo } from "@models/UserInfo";
import { AuthenticationLevel } from "@services/State";
import { setPreferred2FAMethod } from "@services/UserInfo";
import MethodSelectionDialog from "@views/LoginPortal/SecondFactor/MethodSelectionDialog";

const OneTimePasswordMethod = lazy(() => import("@views/LoginPortal/SecondFactor/OneTimePasswordMethod"));
const PushNotificationMethod = lazy(() => import("@views/LoginPortal/SecondFactor/PushNotificationMethod"));
const WebAuthnMethod = lazy(() => import("@views/LoginPortal/SecondFactor/WebAuthnMethod"));
const PasswordMethod = lazy(() => import("@views/LoginPortal/SecondFactor/PasswordMethod"));

export interface Props {
    authenticationLevel: AuthenticationLevel;
    factorKnowledge: boolean;
    userInfo: UserInfo;
    configuration: Configuration;
    duoSelfEnrollment: boolean;

    onMethodChanged: () => void;
    onAuthenticationSuccess: (_redirectURL: string | undefined) => void;
}

const SecondFactorForm = function (props: Props) {
    const { t: translate } = useTranslation();

    const navigate = useRouterNavigate();
    const flowPresent = useFlowPresent();
    const { localStorageMethodAvailable, setLocalStorageMethod } = useLocalStorageMethodContext();
    const { createErrorNotification } = useNotifications();

    const [methodSelectionOpen, setMethodSelectionOpen] = useState(false);
    const stateWebAuthnSupported = browserSupportsWebAuthn();

    const handleMethodSelectionClick = () => {
        setMethodSelectionOpen(true);
    };

    const handleMethodSelected = async (method: SecondFactorMethod) => {
        if (localStorageMethodAvailable) {
            setLocalStorageMethod(method);
        } else {
            await handleMethodSelectedFallback(method);
        }

        setMethodSelectionOpen(false);
        props.onMethodChanged();
    };

    const handleMethodSelectedFallback = async (method: SecondFactorMethod) => {
        try {
            await setPreferred2FAMethod(method);
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("There was an issue updating preferred second factor method"));
        }
    };

    const showMethods = props.factorKnowledge && props.configuration.available_methods.size > 1;

    return (
        <LoginLayout
            id={"second-factor-stage"}
            title={`${translate("Hi")} ${props.userInfo.display_name}`}
            userInfo={props.userInfo}
        >
            {showMethods ? (
                <MethodSelectionDialog
                    open={methodSelectionOpen}
                    methods={props.configuration.available_methods}
                    webauthn={stateWebAuthnSupported}
                    onClose={() => setMethodSelectionOpen(false)}
                    onClick={handleMethodSelected}
                />
            ) : null}
            <div className="flex flex-col items-center justify-center">
                <div className="w-full">
                    <LogoutButton />
                    {flowPresent ? " | " : null}
                    {flowPresent ? <SwitchUserButton /> : null}
                    {showMethods ? " | " : null}
                    {showMethods ? (
                        <Button
                            id={"methods-button"}
                            variant="ghost"
                            className="text-sm tracking-wide text-secondary hover:text-secondary"
                            onClick={handleMethodSelectionClick}
                        >
                            {translate("Methods")}
                        </Button>
                    ) : null}
                </div>
                <div className="my-4 min-w-[300px] rounded-[10px] border border-[#d6d6d6] p-8">
                    <Routes>
                        <Route
                            path={SecondFactorPasswordSubRoute}
                            element={
                                <PasswordMethod
                                    id="password-method"
                                    authenticationLevel={props.authenticationLevel}
                                    onAuthenticationSuccess={props.onAuthenticationSuccess}
                                />
                            }
                        />
                        <Route
                            path={SecondFactorTOTPSubRoute}
                            element={
                                <OneTimePasswordMethod
                                    id={"one-time-password-method"}
                                    authenticationLevel={props.authenticationLevel}
                                    // Whether the user has a TOTP secret registered already
                                    registered={props.userInfo.has_totp}
                                    onRegisterClick={() => {
                                        navigate(`${SettingsRoute}${SettingsTwoFactorAuthenticationSubRoute}`);
                                    }}
                                    onSignInError={(err) => createErrorNotification(err.message)}
                                    onSignInSuccess={props.onAuthenticationSuccess}
                                />
                            }
                        />
                        <Route
                            path={SecondFactorWebAuthnSubRoute}
                            element={
                                <WebAuthnMethod
                                    id={"webauthn-method"}
                                    authenticationLevel={props.authenticationLevel}
                                    // Whether the user has a WebAuthn device registered already
                                    registered={props.userInfo.has_webauthn}
                                    onRegisterClick={() => {
                                        navigate(`${SettingsRoute}${SettingsTwoFactorAuthenticationSubRoute}`);
                                    }}
                                    onSignInError={(err) => createErrorNotification(err.message)}
                                    onSignInSuccess={props.onAuthenticationSuccess}
                                />
                            }
                        />
                        <Route
                            path={SecondFactorPushSubRoute}
                            element={
                                <PushNotificationMethod
                                    id={"push-notification-method"}
                                    authenticationLevel={props.authenticationLevel}
                                    duoSelfEnrollment={props.duoSelfEnrollment}
                                    registered={props.userInfo.has_duo}
                                    onSelectionClick={props.onMethodChanged}
                                    onSignInError={(err) => createErrorNotification(err.message)}
                                    onSignInSuccess={props.onAuthenticationSuccess}
                                />
                            }
                        />
                    </Routes>
                </div>
            </div>
        </LoginLayout>
    );
};

export default SecondFactorForm;
