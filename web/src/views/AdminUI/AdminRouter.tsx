import React, { useEffect } from "react";

import { Route, Routes } from "react-router-dom";

import { IndexRoute } from "@constants/Routes";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { useAutheliaState } from "@hooks/State";
import AdminLayout from "@layouts/AdminLayout";
import { AuthenticationLevel } from "@services/State";
import ClientView from "@views/AdminUI/OpenIDConnect/ClientView";
//import TwoFactorAuthenticationView from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationView";

export interface Props {}

const AdminRouter = function (props: Props) {
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
        <AdminLayout>
            <Routes>
                <Route path={IndexRoute} element={<ClientView />} />
            </Routes>
        </AdminLayout>
    );
};

export default AdminRouter;
