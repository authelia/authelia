import React, { lazy } from "react";

import { Route, Routes } from "react-router-dom";

import { ConsentDecisionSubRoute, ConsentLoginSubRoute } from "@constants/Routes";
import { UserInfo } from "@models/UserInfo";
import { AutheliaState } from "@services/State";
const OpenIDConnectConsentDecisionFormView = lazy(
    () => import("@views/ConsentPortal/OpenIDConnectConsentPortal/OpenIDConnectConsentDecisionFormView"),
);
const OpenIDConnectConsentLoginFormView = lazy(
    () => import("@views/ConsentPortal/OpenIDConnectConsentPortal/OpenIDConnectConsentLoginFormView"),
);

export interface Props {
    userInfo: UserInfo;
    state: AutheliaState;
}

const OpenIDConnectConsentPortal: React.FC<Props> = (props: Props) => {
    return (
        <Routes>
            <Route
                path={ConsentLoginSubRoute}
                element={<OpenIDConnectConsentLoginFormView userInfo={props.userInfo} state={props.state} />}
            />
            <Route
                path={ConsentDecisionSubRoute}
                element={<OpenIDConnectConsentDecisionFormView userInfo={props.userInfo} state={props.state} />}
            />
        </Routes>
    );
};

export default OpenIDConnectConsentPortal;
