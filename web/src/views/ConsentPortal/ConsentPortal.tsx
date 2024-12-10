import React, { Fragment, lazy, useEffect } from "react";

import { useTranslation } from "react-i18next";
import { Route, Routes } from "react-router-dom";

import { ConsentOpenIDSubRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
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

    return (
        <Fragment>
            {state === undefined || userInfo === undefined ? (
                <LoadingPage />
            ) : (
                <ConsentPortalRouter userInfo={userInfo} state={state} />
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
