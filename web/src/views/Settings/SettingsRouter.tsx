import React, { useEffect } from "react";

import { Route, Routes, useNavigate } from "react-router-dom";

import { IndexRoute, SettingsTwoFactorAuthenticationSubRoute } from "@constants/Routes";
import { useAutheliaState } from "@hooks/State";
import { AuthenticationLevel } from "@services/State";
import SettingsView from "@views/Settings/SettingsView";
import TwoFactorAuthenticationView from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationView";

export interface Props {}

const SettingsRouter = function (props: Props) {
    const navigate = useNavigate();
    const [state, fetchState, , fetchStateError] = useAutheliaState();

    // Fetch the state on page load
    useEffect(() => {
        fetchState();
    }, [fetchState]);

    useEffect(() => {
        if (fetchStateError || (state && state.authentication_level < AuthenticationLevel.OneFactor)) {
            navigate(IndexRoute);
        }
    }, [state, fetchStateError, navigate]);

    return (
        <Routes>
            <Route path={IndexRoute} element={<SettingsView />} />
            <Route
                path={SettingsTwoFactorAuthenticationSubRoute}
                element={<TwoFactorAuthenticationView state={state} />}
            />
        </Routes>
    );
};

export default SettingsRouter;
