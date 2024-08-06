import React from "react";

import { Route, Routes } from "react-router-dom";

import { AdminOIDCAuthPoliciesSubRoute, AdminOIDCClientSubRoute, AdminOIDCProviderSubRoute } from "@constants/Routes";
import ClientAuthPoliciesView from "@views/AdminUI/OpenIDConnect/ClientAuthPoliciesView";
import ClientView from "@views/AdminUI/OpenIDConnect/ClientView";
import ProviderView from "@views/AdminUI/OpenIDConnect/ProviderView";

const OIDCRouter = () => {
    return (
        <Routes>
            <Route path={`${AdminOIDCClientSubRoute}`} element={<ClientView />} />
            <Route path={`${AdminOIDCAuthPoliciesSubRoute}`} element={<ClientAuthPoliciesView />} />
            <Route path={`${AdminOIDCProviderSubRoute}`} element={<ProviderView />} />
        </Routes>
    );
};

export default OIDCRouter;
