import { createAction } from 'typesafe-actions';
import {
  FORGOT_PASSWORD_REQUEST,
  FORGOT_PASSWORD_SUCCESS,
  FORGOT_PASSWORD_FAILURE
} from "../../constants";

/*     AUTHENTICATE_REQUEST    */
export const forgotPasswordRequest = createAction(FORGOT_PASSWORD_REQUEST);
export const forgotPasswordSuccess = createAction(FORGOT_PASSWORD_SUCCESS);
export const forgotPasswordFailure = createAction(FORGOT_PASSWORD_FAILURE, resolve => {
  return (error: string) => resolve(error);
});
