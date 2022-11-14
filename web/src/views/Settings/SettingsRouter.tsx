import React from "react";

import { Typography } from "@mui/material";
import { Route, Routes } from "react-router-dom";

import { SettingsRoute, SettingsTwoFactorAuthenticationRoute } from "@constants/Routes";
import SettingsLayout from "@layouts/SettingsLayout";
import TwoFactorAuthenticationView from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationView";

export interface Props {}

const SettingsRouter = function (props: Props) {
    return (
        <Routes>
            <Route>
                path={SettingsRoute} element=
                {
                    <SettingsLayout>
                        <Typography>Portal Placeholder</Typography>
                    </SettingsLayout>
                }
            </Route>
            <Route path={SettingsTwoFactorAuthenticationRoute} element={<TwoFactorAuthenticationView />} />
        </Routes>
    );
};

export default SettingsRouter;
