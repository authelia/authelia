import { useEffect } from "react";

import { Route, Routes } from "react-router";

import {
    IndexRoute,
    SecuritySubRoute,
    SettingsOpenIDConnectSubRoute,
    SettingsTwoFactorAuthenticationSubRoute,
} from "@constants/Routes";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { useAutheliaState } from "@hooks/State";
import SettingsLayout from "@layouts/SettingsLayout";
import { AuthenticationLevel } from "@services/State";
import { getOpenIDConnectLogin } from "@utils/Configuration";
import OpenIDConnectView from "@views/Settings/OpenIDConnect/OpenIDConnectView";
import SecurityView from "@views/Settings/Security/SecurityView";
import SettingsView from "@views/Settings/SettingsView";
import TwoFactorAuthenticationView from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationView";

const SettingsRouter = function () {
    const navigate = useRouterNavigate();
    const [state, fetchState, , fetchStateError] = useAutheliaState();

    const openIDConnectLogin = getOpenIDConnectLogin();

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
                {openIDConnectLogin ? (
                    <Route path={SettingsOpenIDConnectSubRoute} element={<OpenIDConnectView />} />
                ) : null}
            </Routes>
        </SettingsLayout>
    );
};

export default SettingsRouter;
