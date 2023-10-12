import React, { Fragment, useEffect, useState } from "react";

import Grid from "@mui/material/Unstable_Grid2";

import { useNotifications } from "@hooks/NotificationsContext";
import { useUserInfoPOST } from "@hooks/UserInfo";
import { useUserInfoTOTPConfigurationOptional } from "@hooks/UserInfoTOTPConfiguration";
import { useUserWebAuthnCredentials } from "@hooks/WebAuthnCredentials";
import TOTPPanel from "@views/Settings/TwoFactorAuthentication/TOTPPanel";
import WebAuthnCredentialsPanel from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsPanel";

interface Props {}

const TwoFactorAuthSettings = function (props: Props) {
    const [refreshState, setRefreshState] = useState(0);
    const [refreshWebAuthnState, setRefreshWebAuthnState] = useState(0);
    const [refreshTOTPState, setRefreshTOTPState] = useState(0);
    const { createErrorNotification } = useNotifications();
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
        fetchUserInfo();
    }, [fetchUserInfo, refreshState]);

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
        if (fetchUserInfoError) {
            createErrorNotification("There was an issue retrieving user preferences");
        }
    }, [fetchUserInfoError, createErrorNotification]);

    useEffect(() => {
        if (fetchUserTOTPConfigError) {
            createErrorNotification("There was an issue retrieving One Time Password Configuration");
        }
    }, [fetchUserTOTPConfigError, createErrorNotification]);

    useEffect(() => {
        if (fetchUserWebAuthnCredentialsError) {
            createErrorNotification("There was an issue retrieving One Time Password Configuration");
        }
    }, [fetchUserWebAuthnCredentialsError, createErrorNotification]);

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
            </Grid>
        </Fragment>
    );
};

export default TwoFactorAuthSettings;
