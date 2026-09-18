import { ReactNode, useEffect } from "react";

import { Route, Routes } from "react-router";

import {
    IndexRoute,
    SecuritySubRoute,
    SettingsGroupManagementSubRoute,
    SettingsRoute,
    SettingsTwoFactorAuthenticationSubRoute,
    SettingsUserManagementSubRoute,
} from "@constants/Routes";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { useAutheliaState } from "@hooks/State";
import { useUserInfoGET } from "@hooks/UserInfo";
import { useAdminConfigurationGET } from "@hooks/UserManagement";
import SettingsLayout from "@layouts/SettingsLayout";
import { AuthenticationLevel } from "@services/State";
import SecurityView from "@views/Settings/Security/SecurityView";
import SettingsView from "@views/Settings/SettingsView";
import TwoFactorAuthenticationView from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationView";
import GroupManagementView from "@views/Settings/UserManagement/GroupManagementView";
import UserManagementView from "@views/Settings/UserManagement/UserManagementView";

const SettingsRouter = function () {
    const navigate = useRouterNavigate();
    const [state, fetchState, , fetchStateError] = useAutheliaState();

    useEffect(() => {
        fetchState();
    }, [fetchState]);

    useEffect(() => {
        if (fetchStateError || (state && state.authentication_level < AuthenticationLevel.OneFactor)) {
            navigate(IndexRoute);
        }
    }, [state, fetchStateError, navigate]);

    return (
        <SettingsLayout>
            <Routes>
                <Route path={IndexRoute} element={<SettingsView />} />
                <Route path={SecuritySubRoute} element={<SecurityView />} />
                <Route path={SettingsTwoFactorAuthenticationSubRoute} element={<TwoFactorAuthenticationView />} />
                <Route
                    path={SettingsUserManagementSubRoute}
                    element={
                        <AdminRoute>
                            <UserManagementView />
                        </AdminRoute>
                    }
                />
                <Route
                    path={SettingsGroupManagementSubRoute}
                    element={
                        <AdminRoute>
                            <GroupManagementView />
                        </AdminRoute>
                    }
                />
            </Routes>
        </SettingsLayout>
    );
};

interface AdminRouteProps {
    children: ReactNode;
}

const AdminRoute = function (props: AdminRouteProps) {
    const navigate = useRouterNavigate();

    const [userInfo, fetchUserInfo, , fetchUserInfoError] = useUserInfoGET();
    const [adminConfig, fetchAdminConfig, , fetchAdminConfigError] = useAdminConfigurationGET();

    useEffect(() => {
        fetchUserInfo();
        fetchAdminConfig();
    }, [fetchUserInfo, fetchAdminConfig]);

    const loaded =
        (adminConfig !== undefined || fetchAdminConfigError !== undefined) &&
        (userInfo !== undefined || fetchUserInfoError !== undefined);

    const isAdmin =
        !fetchAdminConfigError &&
        !fetchUserInfoError &&
        !!adminConfig?.enabled &&
        !!userInfo?.groups?.includes(adminConfig.admin_group);

    useEffect(() => {
        if (loaded && !isAdmin) {
            navigate(SettingsRoute);
        }
    }, [loaded, isAdmin, navigate]);

    if (!loaded || !isAdmin) {
        return null;
    }

    return <>{props.children}</>;
};

export default SettingsRouter;
