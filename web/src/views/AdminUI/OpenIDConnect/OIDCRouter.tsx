import React from "react";

import { Route, Routes } from "react-router-dom";

import { AdminOIDCAuthPoliciesSubRoute, AdminOIDCClientSubRoute } from "@constants/Routes";
import ClientAuthPoliciesView from "@views/AdminUI/OpenIDConnect/ClientAuthPoliciesView";
import ClientView from "@views/AdminUI/OpenIDConnect/ClientView";

const OIDCRouter = () => {
    return (
        <Routes>
            <Route path={`${AdminOIDCClientSubRoute}`} element={<ClientView />} />
            <Route path={`${AdminOIDCAuthPoliciesSubRoute}`} element={<ClientAuthPoliciesView />} />
        </Routes>
    );
};

export default OIDCRouter;
