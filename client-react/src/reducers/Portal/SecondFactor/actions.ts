import { createAction } from "typesafe-actions";
import {
  LOGOUT_REQUEST,
  LOGOUT_SUCCESS,
  LOGOUT_FAILURE,
  SECURITY_KEY_SIGN,
  SECURITY_KEY_SIGN_SUCCESS,
  SECURITY_KEY_SIGN_FAILURE,
  SET_SECURITY_KEY_SUPPORTED
} from "../../constants";

export const setSecurityKeySupported = createAction(SET_SECURITY_KEY_SUPPORTED, resolve => {
  return (supported: boolean) => resolve(supported);
})

export const securityKeySign = createAction(SECURITY_KEY_SIGN);
export const securityKeySignSuccess = createAction(SECURITY_KEY_SIGN_SUCCESS);
export const securityKeySignFailure = createAction(SECURITY_KEY_SIGN_FAILURE, resolve => {
  return (error: string) => resolve(error);
})

export const logout = createAction(LOGOUT_REQUEST);
export const logoutSuccess = createAction(LOGOUT_SUCCESS);
export const logoutFailure = createAction(LOGOUT_FAILURE, resolve => {
  return (error: string) => resolve(error);
});
