import { FC, lazy } from "react";

import { Route, Routes } from "react-router-dom";

import { ConsentDecisionSubRoute, ConsentOpenIDDeviceAuthorizationSubRoute } from "@constants/Routes";
import { UserInfo } from "@models/UserInfo";
import { AutheliaState } from "@services/State";
const OpenIDConnectConsentDecisionFormView = lazy(() => import("@views/ConsentPortal/OpenIDConnect/DecisionFormView"));
const OpenIDConnectConsentDeviceAuthorizationFormView = lazy(
    () => import("@views/ConsentPortal/OpenIDConnect/DeviceAuthorizationFormView"),
);

export interface Props {
    userInfo?: UserInfo;
    state: AutheliaState;
}

const ConsentPortal: FC<Props> = (props: Props) => {
    return (
        <Routes>
            <Route
                path={ConsentDecisionSubRoute}
                element={<OpenIDConnectConsentDecisionFormView userInfo={props.userInfo} state={props.state} />}
            />
            <Route
                path={ConsentOpenIDDeviceAuthorizationSubRoute}
                element={<OpenIDConnectConsentDeviceAuthorizationFormView state={props.state} />}
            />
        </Routes>
    );
};

export default ConsentPortal;
