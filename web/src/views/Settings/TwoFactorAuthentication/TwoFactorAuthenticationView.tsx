import React, { useEffect, useState } from "react";

import Grid from "@mui/material/Unstable_Grid2";

import { useNotifications } from "@hooks/NotificationsContext";
import { useUserInfoPOST } from "@hooks/UserInfo";
import { useUserWebAuthnDevices } from "@hooks/WebAuthnDevices";
import WebAuthnDevicesPanel from "@views/Settings/TwoFactorAuthentication/WebAuthnDevicesPanel";

interface Props {}

export default function TwoFactorAuthSettings(props: Props) {
    const [refreshState, setRefreshState] = useState(0);
    const { createErrorNotification } = useNotifications();
    const [userInfo, fetchUserInfo, , fetchUserInfoError] = useUserInfoPOST();
    const [userWebAuthnDevices, fetchUserWebAuthnDevices, , fetchUserWebAuthnDevicesError] = useUserWebAuthnDevices();
    const [hasTOTP, setHasTOTP] = useState(false);
    const [hasWebAuthn, setHasWebAuthn] = useState(false);

    const handleRefreshState = () => {
        setRefreshState((refreshState) => refreshState + 1);
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
        fetchUserWebAuthnDevices();
    }, [fetchUserWebAuthnDevices, hasWebAuthn]);

    useEffect(() => {
        if (fetchUserInfoError) {
            createErrorNotification("There was an issue retrieving user preferences");
        }
    }, [fetchUserInfoError, createErrorNotification]);

    useEffect(() => {
        if (fetchUserWebAuthnDevicesError) {
            createErrorNotification("There was an issue retrieving One Time Password Configuration");
        }
    }, [fetchUserWebAuthnDevicesError, createErrorNotification]);

    return (
        <Grid container spacing={2}>
            <Grid xs={12}>
                <WebAuthnDevicesPanel devices={userWebAuthnDevices} handleRefreshState={handleRefreshState} />
            </Grid>
        </Grid>
    );
}
