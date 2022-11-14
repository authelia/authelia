import React from "react";

import { Route, Routes } from "react-router-dom";

import { IndexRoute, SettingsTwoFactorAuthenticationSubRoute } from "@constants/Routes";
import SettingsView from "@views/Settings/SettingsView";
import TwoFactorAuthenticationView from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationView";

export interface Props {}

const SettingsRouter = function (props: Props) {
    return (
        <Routes>
            <Route path={IndexRoute} element={<SettingsView />} />
            <Route path={SettingsTwoFactorAuthenticationSubRoute} element={<TwoFactorAuthenticationView />} />
        </Routes>
    );
};

export default SettingsRouter;
