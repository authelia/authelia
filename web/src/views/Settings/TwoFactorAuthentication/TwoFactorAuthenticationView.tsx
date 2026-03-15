import { useEffect, useState } from "react";

import { useTranslation } from "react-i18next";

import { Card, CardContent } from "@components/UI/Card";
import { useLocalStorageMethodContext } from "@contexts/LocalStorageMethodContext";
import { useConfiguration } from "@hooks/Configuration";
import { useNotifications } from "@hooks/NotificationsContext";
import { useUserInfoPOST } from "@hooks/UserInfo";
import { useUserInfoTOTPConfigurationOptional } from "@hooks/UserInfoTOTPConfiguration";
import { useUserWebAuthnCredentials } from "@hooks/WebAuthnCredentials";
import { SecondFactorMethod } from "@models/Methods";
import OneTimePasswordPanel from "@views/Settings/TwoFactorAuthentication/OneTimePasswordPanel";
import TwoFactorAuthenticationOptionsPanel from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationOptionsPanel";
import WebAuthnCredentialsDisabledPanel from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsDisabledPanel";
import WebAuthnCredentialsPanel from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsPanel";

const TwoFactorAuthenticationView = function () {
    const { t: translate } = useTranslation("settings");

    const [refreshState, setRefreshState] = useState(0);
    const [refreshWebAuthnState, setRefreshWebAuthnState] = useState(0);
    const [refreshTOTPState, setRefreshTOTPState] = useState(0);
    const { createErrorNotification } = useNotifications();

    const [configuration, fetchConfiguration, , fetchConfigurationError] = useConfiguration();
    const [userInfo, fetchUserInfo, , fetchUserInfoError] = useUserInfoPOST();
    const [userTOTPConfig, fetchUserTOTPConfig, , fetchUserTOTPConfigError] = useUserInfoTOTPConfigurationOptional();
    const { localStorageMethodAvailable, setLocalStorageMethod } = useLocalStorageMethodContext();
    const [userWebAuthnCredentials, fetchUserWebAuthnCredentials, , fetchUserWebAuthnCredentialsError] =
        useUserWebAuthnCredentials();

    const hasTOTP = userInfo?.has_totp ?? false;
    const hasWebAuthn = userInfo?.has_webauthn ?? false;

    const handleRefreshWebAuthnState = () => {
        setRefreshState((refreshState) => refreshState + 1);
        setRefreshWebAuthnState((refreshWebAuthnState) => refreshWebAuthnState + 1);
    };

    const handleRefreshTOTPState = () => {
        setRefreshState((refreshState) => refreshState + 1);
        setRefreshTOTPState((refreshTOTPState) => refreshTOTPState + 1);
    };

    const enabledWebAuthn = configuration?.available_methods.has(SecondFactorMethod.WebAuthn);
    const enabledTOTP = configuration?.available_methods.has(SecondFactorMethod.TOTP);

    useEffect(() => {
        fetchConfiguration();
        fetchUserInfo();
    }, [fetchConfiguration, fetchUserInfo, refreshState]);

    useEffect(() => {
        if (localStorageMethodAvailable && configuration?.available_methods.size === 1) {
            setLocalStorageMethod([...configuration.available_methods][0]);
        }
    }, [configuration, localStorageMethodAvailable, setLocalStorageMethod]);

    useEffect(() => {
        if (!enabledTOTP) {
            return;
        }

        fetchUserTOTPConfig();
    }, [enabledTOTP, fetchUserTOTPConfig, hasTOTP, refreshTOTPState]);

    useEffect(() => {
        if (!enabledWebAuthn) {
            return;
        }

        fetchUserWebAuthnCredentials();
    }, [enabledWebAuthn, fetchUserWebAuthnCredentials, hasWebAuthn, refreshWebAuthnState]);

    useEffect(() => {
        if (fetchConfigurationError) {
            createErrorNotification(
                translate("There was an issue retrieving the {{item}}", { item: translate("global configuration") }),
            );
        }
    }, [fetchConfigurationError, createErrorNotification, translate]);

    useEffect(() => {
        if (fetchUserInfoError) {
            createErrorNotification(
                translate("There was an issue retrieving the {{item}}", {
                    item: translate("user preferences"),
                }),
            );
        }
    }, [fetchUserInfoError, createErrorNotification, translate]);

    useEffect(() => {
        if (fetchUserTOTPConfigError) {
            createErrorNotification(
                translate("There was an issue retrieving the {{item}}", {
                    item: translate("One-Time Password configuration"),
                }),
            );
        }
    }, [fetchUserTOTPConfigError, createErrorNotification, translate]);

    useEffect(() => {
        if (fetchUserWebAuthnCredentialsError) {
            createErrorNotification(
                translate("There was an issue retrieving the {{item}}", {
                    item: translate("WebAuthn Credentials"),
                }),
            );
        }
    }, [fetchUserWebAuthnCredentialsError, createErrorNotification, translate]);

    const handleRefreshUserInfo = () => {
        fetchUserInfo();
    };

    const renderSecondFactorDisabled = () => {
        if (!configuration || !userInfo) {
            return false;
        }

        const hasAvailableMethods = enabledTOTP || enabledWebAuthn;

        return !hasAvailableMethods;
    };

    return (
        <div className="grid grid-cols-1 gap-4">
            {renderSecondFactorDisabled() ? (
                <div className="w-full flex justify-center items-center">
                    <Card>
                        <CardContent>
                            <h6 className="text-lg font-semibold p-12 text-center">
                                {translate("There are no protected applications that require a second factor method")}
                            </h6>
                        </CardContent>
                    </Card>
                </div>
            ) : null}
            {!renderSecondFactorDisabled() && enabledTOTP ? (
                <div className="w-full">
                    <OneTimePasswordPanel
                        info={userInfo}
                        config={userTOTPConfig}
                        handleRefreshState={handleRefreshTOTPState}
                    />
                </div>
            ) : null}
            {!renderSecondFactorDisabled() || enabledWebAuthn ? (
                <div className="w-full">
                    {enabledWebAuthn ? (
                        <WebAuthnCredentialsPanel
                            info={userInfo}
                            credentials={userWebAuthnCredentials}
                            handleRefreshState={handleRefreshWebAuthnState}
                        />
                    ) : (
                        <WebAuthnCredentialsDisabledPanel />
                    )}
                </div>
            ) : null}
            {configuration && userInfo ? (
                <div className="w-full">
                    <TwoFactorAuthenticationOptionsPanel
                        config={configuration}
                        info={userInfo}
                        refresh={handleRefreshUserInfo}
                    />
                </div>
            ) : null}
        </div>
    );
};

export default TwoFactorAuthenticationView;
