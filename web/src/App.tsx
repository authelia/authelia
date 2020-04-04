import React, { useState, useEffect } from 'react';
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
import { useConfiguration } from './hooks/Configuration';
import Tracker from "./components/Tracker";
import { useTracking } from "./hooks/Tracking";

const App: React.FC = () => {
    const [notification, setNotification] = useState(null as Notification | null);
    const [configuration, fetchConfig, , fetchConfigError] = useConfiguration();
    const tracker = useTracking(configuration);

    useEffect(() => {
        if (fetchConfigError) {
            console.error(fetchConfigError);
        }
    }, [fetchConfigError]);

    useEffect(() => { fetchConfig() }, [fetchConfig]);

    return (
        <NotificationsContext.Provider value={{ notification, setNotification }} >
            <Router>
                <Tracker tracker={tracker}>
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
                                rememberMe={configuration?.remember_me === true}
                                resetPassword={configuration?.reset_password === true} />
                        </Route>
                        <Route path="/">
                            <Redirect to={FirstFactorRoute}></Redirect>
                        </Route>
                    </Switch>
                </Tracker>
            </Router>
        </NotificationsContext.Provider>
    );
}

export default App;
