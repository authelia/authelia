import { useEffect, useState } from "react";

import { Paper, Typography } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";

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
    const { setLocalStorageMethod, localStorageMethodAvailable } = useLocalStorageMethodContext();
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
        <Grid container spacing={2}>
            {renderSecondFactorDisabled() ? (
                <Grid size={{ xs: 12 }} display="flex" justifyContent="center" alignItems="center">
                    <Paper>
                        <Typography variant={"h6"} sx={{ p: 6 }} text-align="center">
                            {translate("There are no protected applications that require a second factor method")}
                        </Typography>
                    </Paper>
                </Grid>
            ) : null}
            {!renderSecondFactorDisabled() && enabledTOTP ? (
                <Grid size={{ xs: 12 }}>
                    <OneTimePasswordPanel
                        info={userInfo}
                        config={userTOTPConfig}
                        handleRefreshState={handleRefreshTOTPState}
                    />
                </Grid>
            ) : null}
            {!renderSecondFactorDisabled() || enabledWebAuthn ? (
                <Grid size={{ xs: 12 }}>
                    {enabledWebAuthn ? (
                        <WebAuthnCredentialsPanel
                            info={userInfo}
                            credentials={userWebAuthnCredentials}
                            handleRefreshState={handleRefreshWebAuthnState}
                        />
                    ) : (
                        <WebAuthnCredentialsDisabledPanel />
                    )}
                </Grid>
            ) : null}
            {configuration && userInfo ? (
                <Grid size={{ xs: 12 }}>
                    <TwoFactorAuthenticationOptionsPanel
                        config={configuration}
                        info={userInfo}
                        refresh={handleRefreshUserInfo}
                    />
                </Grid>
            ) : null}
        </Grid>
    );
};

export default TwoFactorAuthenticationView;
