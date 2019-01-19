import { createAction } from "typesafe-actions";
import {
  LOGOUT_REQUEST,
  LOGOUT_SUCCESS,
  LOGOUT_FAILURE,
  SECURITY_KEY_SIGN,
  SECURITY_KEY_SIGN_SUCCESS,
  SECURITY_KEY_SIGN_FAILURE,
  SET_SECURITY_KEY_SUPPORTED,
  ONE_TIME_PASSWORD_VERIFICATION_REQUEST,
  ONE_TIME_PASSWORD_VERIFICATION_SUCCESS,
  ONE_TIME_PASSWORD_VERIFICATION_FAILURE
} from "../../constants";

export const setSecurityKeySupported = createAction(SET_SECURITY_KEY_SUPPORTED, resolve => {
  return (supported: boolean) => resolve(supported);
})

export const securityKeySign = createAction(SECURITY_KEY_SIGN);
export const securityKeySignSuccess = createAction(SECURITY_KEY_SIGN_SUCCESS);
export const securityKeySignFailure = createAction(SECURITY_KEY_SIGN_FAILURE, resolve => {
  return (error: string) => resolve(error);
});

export const oneTimePasswordVerification = createAction(ONE_TIME_PASSWORD_VERIFICATION_REQUEST);
export const oneTimePasswordVerificationSuccess = createAction(ONE_TIME_PASSWORD_VERIFICATION_SUCCESS);
export const oneTimePasswordVerificationFailure = createAction(ONE_TIME_PASSWORD_VERIFICATION_FAILURE, resolve => {
  return (err: string) => resolve(err);
});


export const logout = createAction(LOGOUT_REQUEST);
export const logoutSuccess = createAction(LOGOUT_SUCCESS);
export const logoutFailure = createAction(LOGOUT_FAILURE, resolve => {
  return (error: string) => resolve(error);
});
