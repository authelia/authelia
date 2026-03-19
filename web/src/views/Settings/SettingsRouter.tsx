import { useEffect } from "react";

import { Route, Routes } from "react-router-dom";

import {
    IndexRoute,
    SecuritySubRoute,
    SettingsGroupManagementSubRoute,
    SettingsTwoFactorAuthenticationSubRoute,
    SettingsUserManagementSubRoute,
} from "@constants/Routes";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { useAutheliaState } from "@hooks/State";
import SettingsLayout from "@layouts/SettingsLayout";
import { AuthenticationLevel } from "@services/State";
import SecurityView from "@views/Settings/Security/SecurityView";
import SettingsView from "@views/Settings/SettingsView";
import TwoFactorAuthenticationView from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationView";
import UserManagementView from "@views/Settings/UserManagement/UserManagementView";
import GroupManagementView from "@views/Settings/UserManagement/GroupManagementView.tsx";

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
                <Route path={SettingsUserManagementSubRoute} element={<UserManagementView />} />
                <Route path={SettingsGroupManagementSubRoute} element={<GroupManagementView />} />
                <Route path={SettingsTwoFactorAuthenticationSubRoute} element={<TwoFactorAuthenticationView />} />
            </Routes>
        </SettingsLayout>
    );
};

export default SettingsRouter;
