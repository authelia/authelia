import React, { Fragment, useEffect, useState } from "react";

import { Paper, Typography } from "@mui/material";
import Grid from "@mui/material/Grid2";
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
import WebAuthnCredentialsPanel from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsPanel";

interface Props {}

const TwoFactorAuthenticationView = function (props: Props) {
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
    const [hasTOTP, setHasTOTP] = useState(false);
    const [hasWebAuthn, setHasWebAuthn] = useState(false);

    const handleRefreshWebAuthnState = () => {
        setRefreshState((refreshState) => refreshState + 1);
        setRefreshWebAuthnState((refreshWebAuthnState) => refreshWebAuthnState + 1);
    };

    const handleRefreshTOTPState = () => {
        setRefreshState((refreshState) => refreshState + 1);
        setRefreshTOTPState((refreshTOTPState) => refreshTOTPState + 1);
    };

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
        if (userInfo === undefined) {
            return;
        }

        if (userInfo.has_webauthn !== hasWebAuthn) {
            setHasWebAuthn(userInfo.has_webauthn);
        }

        if (userInfo.has_totp !== hasTOTP) {
            setHasTOTP(userInfo.has_totp);
        }
    }, [hasTOTP, hasWebAuthn, userInfo]);

    useEffect(() => {
        fetchUserTOTPConfig();
    }, [fetchUserTOTPConfig, hasTOTP, refreshTOTPState]);

    useEffect(() => {
        fetchUserWebAuthnCredentials();
    }, [fetchUserWebAuthnCredentials, hasWebAuthn, refreshWebAuthnState]);

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

        const hasAvailableMethods =
            configuration.available_methods.has(SecondFactorMethod.TOTP) ||
            configuration.available_methods.has(SecondFactorMethod.WebAuthn);

        return !hasAvailableMethods;
    };

    return (
        <Fragment>
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
                {configuration?.available_methods.has(SecondFactorMethod.TOTP) ? (
                    <Grid size={{ xs: 12 }}>
                        <OneTimePasswordPanel
                            info={userInfo}
                            config={userTOTPConfig}
                            handleRefreshState={handleRefreshTOTPState}
                        />
                    </Grid>
                ) : null}
                {configuration?.available_methods.has(SecondFactorMethod.WebAuthn) ? (
                    <Grid size={{ xs: 12 }}>
                        <WebAuthnCredentialsPanel
                            info={userInfo}
                            credentials={userWebAuthnCredentials}
                            handleRefreshState={handleRefreshWebAuthnState}
                        />
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
                ) : undefined}
            </Grid>
        </Fragment>
    );
};

export default TwoFactorAuthenticationView;
