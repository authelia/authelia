import React, { useEffect } from "react";

import { Route, Routes } from "react-router-dom";

import { IndexRoute, SettingsTestSubRoute, SettingsTwoFactorAuthenticationSubRoute } from "@constants/Routes";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { useAutheliaState } from "@hooks/State";
import SettingsLayout from "@layouts/SettingsLayout";
import { AuthenticationLevel } from "@services/State";
import SettingsView from "@views/Settings/SettingsView";
import TestView from "@views/Settings/Test/TestView";
import TwoFactorAuthenticationView from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationView";

export interface Props {}

const SettingsRouter = function (props: Props) {
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
                <Route path={SettingsTwoFactorAuthenticationSubRoute} element={<TwoFactorAuthenticationView />} />
                <Route path={SettingsTestSubRoute} element={<TestView />} />
            </Routes>
        </SettingsLayout>
    );
};

export default SettingsRouter;
