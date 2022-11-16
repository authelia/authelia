import React, { useEffect, useState } from "react";

import { Grid } from "@mui/material";

import { useConfiguration } from "@hooks/Configuration";
import { useNotifications } from "@hooks/NotificationsContext";
import { useUserInfoPOST } from "@hooks/UserInfo";
import SettingsLayout from "@layouts/SettingsLayout";
import { Configuration } from "@models/Configuration";
import { UserInfo } from "@models/UserInfo";
import { AutheliaState, AuthenticationLevel } from "@services/State";

import TOTP from "./TOTP";
import WebauthnDevices from "./WebauthnDevices";

interface Props {
    state: AutheliaState;
}

interface MethodProps extends Props {
    configuration: Configuration;
    userInfo: UserInfo;
}

export default function TwoFactorAuthenticationView(props: Props) {
    const { createErrorNotification } = useNotifications();
    const [userInfo, fetchUserInfo, , fetchUserInfoError] = useUserInfoPOST();
    const [configuration, fetchConfiguration, , fetchConfigurationError] = useConfiguration();

    const [methodProps, setMethodProps] = useState(null as MethodProps | null);

    // Fetch preferences and configuration on load
    useEffect(() => {
        if (props.state && props.state.authentication_level >= AuthenticationLevel.OneFactor) {
            fetchUserInfo();
            fetchConfiguration();
        }
    }, [props.state, fetchUserInfo, fetchConfiguration]);

    useEffect(() => {
        let mProps: MethodProps = {
            state: props.state,
            configuration: configuration,
            userInfo: userInfo,
        };
        if (Object.values(mProps).every((x) => !!x)) {
            setMethodProps(mProps);
            console.log(mProps);
        }
    }, [props.state, configuration, userInfo]);

    // Display an error when configuration fetching fails
    useEffect(() => {
        if (fetchConfigurationError) {
            createErrorNotification("There was an issue retrieving global configuration");
        }
    }, [fetchConfigurationError, createErrorNotification]);

    // Display an error when preferences fetching fails
    useEffect(() => {
        if (fetchUserInfoError) {
            createErrorNotification("There was an issue retrieving user preferences");
        }
    }, [fetchUserInfoError, createErrorNotification]);

    return (
        <SettingsLayout>
            <Grid container spacing={2}>
                {!!methodProps && (
                    <>
                        <Grid item xs={12}>
                            <WebauthnDevices {...methodProps} />
                        </Grid>
                        <Grid item xs={12}>
                            <TOTP {...methodProps} />
                        </Grid>
                    </>
                )}
            </Grid>
        </SettingsLayout>
    );
}
