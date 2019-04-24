import { createAction } from 'typesafe-actions';
import {
  AUTHENTICATE_REQUEST,
  AUTHENTICATE_SUCCESS,
  AUTHENTICATE_FAILURE,
  FIRST_FACTOR_SET_USERNAME,
  FIRST_FACTOR_SET_PASSWORD
} from "../../constants";

/*     AUTHENTICATE_REQUEST    */
export const authenticate = createAction(AUTHENTICATE_REQUEST);
export const authenticateSuccess = createAction(AUTHENTICATE_SUCCESS);
export const authenticateFailure = createAction(AUTHENTICATE_FAILURE, resolve => {
  return (error: string) => resolve(error);
});


export const setUsername = createAction(FIRST_FACTOR_SET_USERNAME, resolve => {
  return (username: string) => resolve(username);
});
export const setPassword = createAction(FIRST_FACTOR_SET_PASSWORD, resolve => {
  return (password: string) => resolve(password);
});