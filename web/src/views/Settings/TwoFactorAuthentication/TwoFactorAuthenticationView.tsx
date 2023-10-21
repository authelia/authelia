import React, { Fragment, useEffect, useState } from "react";

import Grid from "@mui/material/Unstable_Grid2/Grid2";
import { useTranslation } from "react-i18next";

import { useConfiguration } from "@hooks/Configuration";
import { useNotifications } from "@hooks/NotificationsContext";
import { useUserInfoPOST } from "@hooks/UserInfo";
import { useUserInfoTOTPConfigurationOptional } from "@hooks/UserInfoTOTPConfiguration";
import { useUserWebAuthnCredentials } from "@hooks/WebAuthnCredentials";
import TOTPPanel from "@views/Settings/TwoFactorAuthentication/TOTPPanel";
import TwoFactorAuthenticationOptionsPanel from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationOptionsPanel";
import WebAuthnCredentialsPanel from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsPanel";

interface Props {}

const TwoFactorAuthenticationView = function (props: Props) {
    const { t: translate } = useTranslation();

    const [refreshState, setRefreshState] = useState(0);
    const [refreshWebAuthnState, setRefreshWebAuthnState] = useState(0);
    const [refreshTOTPState, setRefreshTOTPState] = useState(0);
    const { createErrorNotification } = useNotifications();

    const [configuration, fetchConfiguration, , fetchConfigurationError] = useConfiguration();
    const [userInfo, fetchUserInfo, , fetchUserInfoError] = useUserInfoPOST();
    const [userTOTPConfig, fetchUserTOTPConfig, , fetchUserTOTPConfigError] = useUserInfoTOTPConfigurationOptional();
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
            createErrorNotification(translate("There was an issue retrieving the global configuration"));
        }
    }, [fetchConfigurationError, createErrorNotification, translate]);

    useEffect(() => {
        if (fetchUserInfoError) {
            createErrorNotification(translate("There was an issue retrieving the user preferences"));
        }
    }, [fetchUserInfoError, createErrorNotification, translate]);

    useEffect(() => {
        if (fetchUserTOTPConfigError) {
            createErrorNotification(translate("There was an issue retrieving One-Time Password configuration"));
        }
    }, [fetchUserTOTPConfigError, createErrorNotification, translate]);

    useEffect(() => {
        if (fetchUserWebAuthnCredentialsError) {
            createErrorNotification(translate("There was an issue retrieving the WebAuthn credentials"));
        }
    }, [fetchUserWebAuthnCredentialsError, createErrorNotification, translate]);

    const handleRefreshUserInfo = () => {
        fetchUserInfo();
    };

    useEffect(() => {
        console.table(userInfo);
        console.table(configuration);
    }, [configuration, userInfo]);

    return (
        <Fragment>
            <Grid container spacing={2}>
                <Grid xs={12}>
                    <TOTPPanel config={userTOTPConfig} handleRefreshState={handleRefreshTOTPState} />
                </Grid>
                <Grid xs={12}>
                    <WebAuthnCredentialsPanel
                        credentials={userWebAuthnCredentials}
                        handleRefreshState={handleRefreshWebAuthnState}
                    />
                </Grid>
                {configuration && userInfo ? (
                    <Grid xs={12}>
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
