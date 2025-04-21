import React, { Fragment, lazy, useEffect } from "react";

import { Button } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";
import { Route, Routes } from "react-router-dom";

import { ConsentOpenIDSubRoute, LogoutRoute as SignOutRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { useAutheliaState } from "@hooks/State";
import { useUserInfoGET } from "@hooks/UserInfo";
import { UserInfo } from "@models/UserInfo";
import { AutheliaState } from "@services/State";
import LoadingPage from "@views/LoadingPage/LoadingPage";

const OpenIDConnectConsentPortal = lazy(
    () => import("@views/ConsentPortal/OpenIDConnectConsentPortal/OpenIDConnectConsentPortal"),
);

export interface Props {}

const ConsentPortal: React.FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation();
    const navigate = useRouterNavigate();

    const [userInfo, fetchUserInfo, , fetchUserInfoError] = useUserInfoGET();
    const [state, fetchState, , fetchStateError] = useAutheliaState();

    const { createErrorNotification, resetNotification } = useNotifications();

    useEffect(() => {
        fetchState();
        fetchUserInfo();
    }, [fetchState, fetchUserInfo]);

    useEffect(() => {
        if (fetchUserInfoError) {
            createErrorNotification(translate("There was an issue retrieving user preferences"));
        }
    }, [fetchUserInfoError, resetNotification, createErrorNotification, translate]);

    useEffect(() => {
        if (fetchStateError) {
            createErrorNotification(translate("There was an issue retrieving the current user state"));
        }
    }, [fetchStateError, createErrorNotification, translate]);

    const handleLogoutClick = () => {
        navigate(SignOutRoute);
    };

    return (
        <Fragment>
            {state === undefined || userInfo === undefined ? (
                <LoadingPage />
            ) : (
                <Fragment>
                    <Grid container direction={"row"}>
                        <Grid size={{ xs: 12 }}>
                            <Grid container direction={"column"} justifyContent={"center"} alignItems={"center"}>
                                <Grid size={{ xs: 12 }}>
                                    <Button id={"logout-button"} color={"secondary"} onClick={handleLogoutClick}>
                                        {translate("Logout")}
                                    </Button>
                                </Grid>
                            </Grid>
                        </Grid>
                        <Grid size={{ xs: 12 }}>
                            <ConsentPortalRouter userInfo={userInfo} state={state} />
                        </Grid>
                    </Grid>
                </Fragment>
            )}
        </Fragment>
    );
};

interface RouterProps {
    userInfo: UserInfo;
    state: AutheliaState;
}

const ConsentPortalRouter: React.FC<RouterProps> = (props: RouterProps) => {
    return (
        <Routes>
            <Route
                path={`${ConsentOpenIDSubRoute}/*`}
                element={<OpenIDConnectConsentPortal userInfo={props.userInfo} state={props.state} />}
            />
        </Routes>
    );
};

export default ConsentPortal;
