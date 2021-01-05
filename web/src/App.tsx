import React, { useState } from "react";

import { config as faConfig } from "@fortawesome/fontawesome-svg-core";
import { CssBaseline, ThemeProvider } from "@material-ui/core";
import { BrowserRouter as Router, Route, Switch, Redirect } from "react-router-dom";

import NotificationBar from "./components/NotificationBar";
import NotificationsContext from "./hooks/NotificationsContext";
import { Notification } from "./models/Notifications";
import {
    FirstFactorRoute,
    ResetPasswordStep2Route,
    ResetPasswordStep1Route,
    RegisterSecurityKeyRoute,
    RegisterOneTimePasswordRoute,
    LogoutRoute,
} from "./Routes";
import * as themes from "./themes";
import { getBasePath } from "./utils/BasePath";
import { getRememberMe, getResetPassword, getTheme } from "./utils/Configuration";
import RegisterOneTimePassword from "./views/DeviceRegistration/RegisterOneTimePassword";
import RegisterSecurityKey from "./views/DeviceRegistration/RegisterSecurityKey";
import LoginPortal from "./views/LoginPortal/LoginPortal";
import SignOut from "./views/LoginPortal/SignOut/SignOut";
import ResetPasswordStep1 from "./views/ResetPassword/ResetPasswordStep1";
import ResetPasswordStep2 from "./views/ResetPassword/ResetPasswordStep2";

import "@fortawesome/fontawesome-svg-core/styles.css";

faConfig.autoAddCss = false;

function Theme() {
    switch (getTheme()) {
        case "dark":
            return themes.Dark;
        case "grey":
            return themes.Grey;
        case "custom":
            return themes.Custom;
        default:
            return themes.Light;
    }
}

const App: React.FC = () => {
    const [notification, setNotification] = useState(null as Notification | null);

    return (
        <ThemeProvider theme={Theme()}>
            <CssBaseline />
            <NotificationsContext.Provider value={{ notification, setNotification }}>
                <Router basename={getBasePath()}>
                    <NotificationBar onClose={() => setNotification(null)} />
                    <Switch>
                        <Route path={ResetPasswordStep1Route} exact>
                            <ResetPasswordStep1 />
                        </Route>
                        <Route path={ResetPasswordStep2Route} exact>
                            <ResetPasswordStep2 />
                        </Route>
                        <Route path={RegisterSecurityKeyRoute} exact>
                            <RegisterSecurityKey />
                        </Route>
                        <Route path={RegisterOneTimePasswordRoute} exact>
                            <RegisterOneTimePassword />
                        </Route>
                        <Route path={LogoutRoute} exact>
                            <SignOut />
                        </Route>
                        <Route path={FirstFactorRoute}>
                            <LoginPortal rememberMe={getRememberMe()} resetPassword={getResetPassword()} />
                        </Route>
                        <Route path="/">
                            <Redirect to={FirstFactorRoute} />
                        </Route>
                    </Switch>
                </Router>
            </NotificationsContext.Provider>
        </ThemeProvider>
    );
};

export default App;
