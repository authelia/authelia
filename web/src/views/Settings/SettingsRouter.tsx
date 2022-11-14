import React from "react";

import { Typography } from "@mui/material";
import { Route, Routes } from "react-router-dom";

import { IndexRoute, SettingsTwoFactorAuthenticationSubRoute } from "@constants/Routes";
import SettingsLayout from "@layouts/SettingsLayout";
import TwoFactorAuthenticationView from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationView";

export interface Props {}

const SettingsRouter = function (props: Props) {
    return (
        <Routes>
            <Route>
                path={IndexRoute} element=
                {
                    <SettingsLayout>
                        <Typography>Portal Placeholder</Typography>
                    </SettingsLayout>
                }
            </Route>
            <Route path={SettingsTwoFactorAuthenticationSubRoute} element={<TwoFactorAuthenticationView />} />
        </Routes>
    );
};

export default SettingsRouter;
