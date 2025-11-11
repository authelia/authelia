import React, { Fragment, lazy, useEffect } from "react";

import { useTranslation } from "react-i18next";
import { Route, Routes } from "react-router-dom";

import { ConsentCompletionSubRoute, ConsentOpenIDSubRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import { useAutheliaState } from "@hooks/State";
import { useUserInfoGET } from "@hooks/UserInfo";
import { UserInfo } from "@models/UserInfo";
import { AutheliaState, AuthenticationLevel } from "@services/State";
import LoadingPage from "@views/LoadingPage/LoadingPage";

const OpenIDConnect = lazy(() => import("@views/ConsentPortal/OpenIDConnect/ConsentPortal"));
const CompletionView = lazy(() => import("@views/ConsentPortal/CompletionView"));

export interface Props {}

const ConsentPortal: React.FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation();

    const [userInfo, fetchUserInfo, , fetchUserInfoError] = useUserInfoGET();
    const [state, fetchState, , fetchStateError] = useAutheliaState();
    const loading = !state || (state.authentication_level >= AuthenticationLevel.OneFactor && !userInfo);
    const { createErrorNotification, resetNotification } = useNotifications();

    useEffect(() => {
        fetchState();
    }, [fetchState, fetchUserInfo]);

    useEffect(() => {
        if (state) {
            if (state.authentication_level >= AuthenticationLevel.OneFactor) {
                fetchUserInfo();
            }
        }
    }, [state, fetchUserInfo]);

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
            {loading || !state ? <LoadingPage /> : <ConsentPortalRouter userInfo={userInfo} state={state} />}
        </Fragment>
    );
};

interface RouterProps {
    userInfo?: UserInfo;
    state: AutheliaState;
}

const ConsentPortalRouter: React.FC<RouterProps> = (props: RouterProps) => {
    return (
        <Routes>
            <Route
                path={`${ConsentOpenIDSubRoute}/*`}
                element={<OpenIDConnect userInfo={props.userInfo} state={props.state} />}
            />
            <Route
                path={ConsentCompletionSubRoute}
                element={<CompletionView userInfo={props.userInfo} state={props.state} />}
            />
        </Routes>
    );
};

export default ConsentPortal;
