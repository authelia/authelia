// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useEffect } from "react";

import { Route, Routes } from "react-router";

import { IndexRoute, SecuritySubRoute, SettingsTwoFactorAuthenticationSubRoute } from "@constants/Routes";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { useAutheliaState } from "@hooks/State";
import { useUserSessionElevation } from "@hooks/UserSessionElevation";
import SettingsLayout from "@layouts/SettingsLayout";
import { AuthenticationLevel } from "@services/State";
import SecurityView from "@views/Settings/Security/SecurityView";
import SettingsView from "@views/Settings/SettingsView";
import TwoFactorAuthenticationView from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationView";

const SettingsRouter = function () {
    const navigate = useRouterNavigate();
    const [state, fetchState, , fetchStateError] = useAutheliaState();
    const [elevation, fetchElevation, , fetchElevationError] = useUserSessionElevation();

    const oneFactor = state?.authentication_level === AuthenticationLevel.OneFactor;
    const allowed =
        state !== undefined &&
        state.authentication_level >= AuthenticationLevel.OneFactor &&
        (!oneFactor || (elevation !== undefined && !elevation.require_second_factor));

    useEffect(() => {
        fetchState();
    }, [fetchState]);

    useEffect(() => {
        if (oneFactor) {
            fetchElevation();
        }
    }, [oneFactor, fetchElevation]);

    useEffect(() => {
        if (fetchStateError || (state && state.authentication_level < AuthenticationLevel.OneFactor)) {
            navigate(IndexRoute);
        } else if (oneFactor && (fetchElevationError || elevation?.require_second_factor)) {
            navigate(IndexRoute);
        }
    }, [state, fetchStateError, oneFactor, elevation, fetchElevationError, navigate]);

    if (!allowed) {
        return null;
    }

    return (
        <SettingsLayout>
            <Routes>
                <Route path={IndexRoute} element={<SettingsView />} />
                <Route path={SecuritySubRoute} element={<SecurityView />} />
                <Route path={SettingsTwoFactorAuthenticationSubRoute} element={<TwoFactorAuthenticationView />} />
            </Routes>
        </SettingsLayout>
    );
};

export default SettingsRouter;
