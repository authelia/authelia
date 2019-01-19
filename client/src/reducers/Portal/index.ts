import { combineReducers } from 'redux';

import FirstFactorReducer from './FirstFactor/reducer';
import SecondFactorReducer from './SecondFactor/reducer';
import OneTimePasswordRegistrationReducer from './OneTimePasswordRegistration/reducer';
import SecurityKeyRegistrationReducer from './SecurityKeyRegistration/reducer';
import AuthenticationReducer from './Authentication/reducer';
import ForgotPasswordReducer from './ForgotPassword/reducer';
import ResetPasswordReducer from './ResetPassword/reducer';

import { connectRouter } from 'connected-react-router'
import { History } from 'history';

function reducer(history: History) {
  return combineReducers({
    router: connectRouter(history),
    authentication: AuthenticationReducer,
    firstFactor: FirstFactorReducer,
    secondFactor: SecondFactorReducer,
    oneTimePasswordRegistration: OneTimePasswordRegistrationReducer,
    securityKeyRegistration: SecurityKeyRegistrationReducer,
    forgotPassword: ForgotPasswordReducer,
    resetPassword: ResetPasswordReducer,
  });
}


export default reducer;