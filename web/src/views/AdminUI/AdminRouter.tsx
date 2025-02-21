import React, { useEffect } from "react";

import { Route, Routes } from "react-router-dom";

import { AdminOIDCSubRoute, IndexRoute } from "@constants/Routes";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { useAutheliaState } from "@hooks/State";
import AdminLayout from "@layouts/AdminLayout";
import { AuthenticationLevel } from "@services/State";
import AdminView from "@views/AdminUI/AdminView";
import OIDCRouter from "@views/AdminUI/OpenIDConnect/OIDCRouter";

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
                <Route path={`${AdminOIDCSubRoute}/*`} element={<OIDCRouter />} />
                <Route path={IndexRoute} element={<AdminView />} />
            </Routes>
        </AdminLayout>
    );
};

export default AdminRouter;
