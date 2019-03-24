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
  ONE_TIME_PASSWORD_VERIFICATION_FAILURE,
  GET_PREFERED_METHOD,
  GET_PREFERED_METHOD_SUCCESS,
  GET_PREFERED_METHOD_FAILURE,
  SET_PREFERED_METHOD,
  SET_PREFERED_METHOD_FAILURE,
  SET_PREFERED_METHOD_SUCCESS,
  SET_USE_ANOTHER_METHOD,
  TRIGGER_DUO_PUSH_AUTH,
  TRIGGER_DUO_PUSH_AUTH_SUCCESS,
  TRIGGER_DUO_PUSH_AUTH_FAILURE,
  GET_AVAILABLE_METHODS,
  GET_AVAILABLE_METHODS_SUCCESS,
  GET_AVAILABLE_METHODS_FAILURE
} from "../../constants";
import Method2FA from "../../../types/Method2FA";

export const setSecurityKeySupported = createAction(SET_SECURITY_KEY_SUPPORTED, resolve => {
  return (supported: boolean) => resolve(supported);
});

export const setUseAnotherMethod = createAction(SET_USE_ANOTHER_METHOD, resolve => {
  return (useAnotherMethod: boolean) => resolve(useAnotherMethod);
});


export const getAvailbleMethods = createAction(GET_AVAILABLE_METHODS);
export const getAvailbleMethodsSuccess = createAction(GET_AVAILABLE_METHODS_SUCCESS, resolve => {
  return (methods: Method2FA[]) => resolve(methods);
});
export const getAvailbleMethodsFailure = createAction(GET_AVAILABLE_METHODS_FAILURE, resolve => {
  return (err: string) => resolve(err);
});


export const getPreferedMethod = createAction(GET_PREFERED_METHOD);
export const getPreferedMethodSuccess = createAction(GET_PREFERED_METHOD_SUCCESS, resolve => {
  return (method: Method2FA) => resolve(method);
});
export const getPreferedMethodFailure = createAction(GET_PREFERED_METHOD_FAILURE, resolve => {
  return (err: string) => resolve(err);
});


export const setPreferedMethod = createAction(SET_PREFERED_METHOD);
export const setPreferedMethodSuccess = createAction(SET_PREFERED_METHOD_SUCCESS);
export const setPreferedMethodFailure = createAction(SET_PREFERED_METHOD_FAILURE, resolve => {
  return (err: string) => resolve(err);
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


export const triggerDuoPushAuth = createAction(TRIGGER_DUO_PUSH_AUTH);
export const triggerDuoPushAuthSuccess = createAction(TRIGGER_DUO_PUSH_AUTH_SUCCESS);
export const triggerDuoPushAuthFailure = createAction(TRIGGER_DUO_PUSH_AUTH_FAILURE, resolve => {
  return (err: string) => resolve(err);
});


export const logout = createAction(LOGOUT_REQUEST);
export const logoutSuccess = createAction(LOGOUT_SUCCESS);
export const logoutFailure = createAction(LOGOUT_FAILURE, resolve => {
  return (error: string) => resolve(error);
});
