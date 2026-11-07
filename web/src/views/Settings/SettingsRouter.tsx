// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useEffect } from "react";

import { Route, Routes } from "react-router";

import {
    IndexRoute,
    SecuritySubRoute,
    SettingsExternalIdentitySubRoute,
    SettingsTwoFactorAuthenticationSubRoute,
} from "@constants/Routes";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { useAutheliaState } from "@hooks/State";
import SettingsLayout from "@layouts/SettingsLayout";
import { AuthenticationLevel } from "@services/State";
import { getExternalIdentityLogin } from "@utils/Configuration";
import ExternalIdentityView from "@views/Settings/ExternalIdentity/ExternalIdentityView";
import SecurityView from "@views/Settings/Security/SecurityView";
import SettingsView from "@views/Settings/SettingsView";
import TwoFactorAuthenticationView from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationView";

const SettingsRouter = function () {
    const navigate = useRouterNavigate();
    const [state, fetchState, , fetchStateError] = useAutheliaState();

    const externalIdentityLogin = getExternalIdentityLogin();

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
                {externalIdentityLogin ? (
                    <Route path={SettingsExternalIdentitySubRoute} element={<ExternalIdentityView />} />
                ) : null}
            </Routes>
        </SettingsLayout>
    );
};

export default SettingsRouter;
