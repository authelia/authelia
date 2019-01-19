import { createAction } from 'typesafe-actions';
import { RESET_PASSWORD_REQUEST, RESET_PASSWORD_SUCCESS, RESET_PASSWORD_FAILURE } from "../../constants";

/*     AUTHENTICATE_REQUEST    */
export const resetPasswordRequest = createAction(RESET_PASSWORD_REQUEST);
export const resetPasswordSuccess = createAction(RESET_PASSWORD_SUCCESS);
export const resetPasswordFailure = createAction(RESET_PASSWORD_FAILURE, resolve => {
  return (error: string) => resolve(error);
});
