import React, { useState } from 'react';
import {
    BrowserRouter as Router, Route, Switch, Redirect
} from "react-router-dom";
import ResetPasswordStep1 from './views/ResetPassword/ResetPasswordStep1';
import ResetPasswordStep2 from './views/ResetPassword/ResetPasswordStep2';
import RegisterSecurityKey from './views/DeviceRegistration/RegisterSecurityKey';
import RegisterOneTimePassword from './views/DeviceRegistration/RegisterOneTimePassword';
import {
    FirstFactorRoute, ResetPasswordStep2Route,
    ResetPasswordStep1Route, RegisterSecurityKeyRoute,
    RegisterOneTimePasswordRoute,
    LogoutRoute,
} from "./Routes";
import LoginPortal from './views/LoginPortal/LoginPortal';
import NotificationsContext from './hooks/NotificationsContext';
import { Notification } from './models/Notifications';
import NotificationBar from './components/NotificationBar';
import SignOut from './views/LoginPortal/SignOut/SignOut';
import { getRememberMe, getResetPassword } from './utils/Configuration';
import '@fortawesome/fontawesome-svg-core/styles.css'
import { config as faConfig } from '@fortawesome/fontawesome-svg-core';
import { getBasePath } from './utils/BasePath';

faConfig.autoAddCss = false;

const App: React.FC = () => {
    const [notification, setNotification] = useState(null as Notification | null);

    return (
        <NotificationsContext.Provider value={{ notification, setNotification }} >
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
                        <LoginPortal
                            rememberMe={getRememberMe()}
                            resetPassword={getResetPassword()} />
                    </Route>
                    <Route path="/">
                        <Redirect to={FirstFactorRoute} />
                    </Route>
                </Switch>
            </Router>
        </NotificationsContext.Provider>
    );
}

export default App;
