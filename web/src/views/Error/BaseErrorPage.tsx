import React, { useEffect, useState } from "react";

import { Button, Grid, Theme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";
import { Path, useNavigate, useSearchParams } from "react-router-dom";

import { IndexRoute, LogoutRoute } from "@constants/Routes";
import { ErrorCode, RedirectionURL } from "@constants/SearchParams";
import { useNotifications } from "@hooks/NotificationsContext";
import { useAutheliaState } from "@hooks/State";
import { useUserInfoPOST } from "@hooks/UserInfo";
import LoginLayout from "@layouts/LoginLayout";
import { Errors } from "@models/Errors";
import { AuthenticationLevel } from "@services/State";
import GenericError from "@views/Error/GenericError";
import ForbiddenError from "@views/Error/NamedErrors/ForbiddenError";
import { ComponentOrLoading } from "@views/Generic/ComponentOrLoading";

const BaseErrorView = function () {
    const styles = useStyles();
    const navigate = useNavigate();
    const { t: translate } = useTranslation();
    const [navigationRoute] = useState({ pathname: LogoutRoute } as Partial<Path>);
    const { createErrorNotification } = useNotifications();
    const [state, fetchState, , fetchStateError] = useAutheliaState();
    const [searchParams] = useSearchParams();
    const [searchParamsOverride] = useState(new URLSearchParams());
    const [userInfo, fetchUserInfo, , fetchUserInfoError] = useUserInfoPOST();

    // Fetch the state when portal is mounted.
    useEffect(() => {
        fetchState();
    }, [fetchState]);

    // Fetch preferences and configuration when user is authenticated.
    useEffect(() => {
        if (state && state.authentication_level >= AuthenticationLevel.OneFactor) {
            fetchUserInfo();
        }
    }, [state, fetchUserInfo]);

    // Display an error when state fetching fails
    useEffect(() => {
        if (fetchStateError) {
            createErrorNotification("There was an issue fetching the current user state");
        }
    }, [fetchStateError, createErrorNotification]);

    useEffect(() => {
        switch (searchParams.get(ErrorCode)) {
            case Errors.forbidden: {
                if (searchParams.has(RedirectionURL)) {
                    searchParamsOverride.set(RedirectionURL, searchParams.get(RedirectionURL) as string);
                    navigationRoute.search = searchParamsOverride.toString();
                }
                if (state?.authentication_level === AuthenticationLevel.Unauthenticated) {
                    navigationRoute.pathname = IndexRoute;
                    navigate(navigationRoute);
                }
                break;
            }
            default: {
                break;
            }
        }
    }, [state, searchParams, navigationRoute, navigate, searchParamsOverride]);

    // Display an error when user information fetching fails
    useEffect(() => {
        if (fetchUserInfoError) {
            createErrorNotification("There was an issue retrieving user information");
        }
    }, [fetchUserInfoError, createErrorNotification]);

    const handleErrorCodeJSX = () => {
        switch (searchParams.get(ErrorCode)) {
            case Errors.forbidden: {
                return <ForbiddenError />;
            }
            default: {
                return <GenericError />;
            }
        }
    };

    const handleLogoutClick = () => {
        navigate(navigationRoute);
    };

    const infoLoaded = userInfo !== undefined;

    return (
        <ComponentOrLoading ready={infoLoaded}>
            <LoginLayout id="base-error-stage" title={`${translate("Hi")} ${userInfo?.display_name || ""}`}>
                <Grid container>
                    <Grid item xs={12}>
                        <Button color="secondary" onClick={handleLogoutClick} id="logout-button">
                            {translate("Logout")}
                        </Button>
                    </Grid>
                    <Grid item xs={12} className={styles.mainContainer}>
                        {handleErrorCodeJSX()}
                    </Grid>
                </Grid>
            </LoginLayout>
        </ComponentOrLoading>
    );
};

export default BaseErrorView;

const useStyles = makeStyles((theme: Theme) => ({
    mainContainer: {
        border: "1px solid #d6d6d6",
        borderRadius: "10px",
        padding: theme.spacing(4),
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
}));
