import { combineReducers } from 'redux';

import FirstFactorReducer from './FirstFactor/reducer';
import SecondFactorReducer from './SecondFactor/reducer';
import OneTimePasswordRegistrationReducer from './OneTimePasswordRegistration/reducer';
import SecurityKeyRegistrationReducer from './SecurityKeyRegistration/reducer';

export default combineReducers({
  firstFactor: FirstFactorReducer,
  secondFactor: SecondFactorReducer,
  oneTimePasswordRegistration: OneTimePasswordRegistrationReducer,
  securityKeyRegistration: SecurityKeyRegistrationReducer,
});