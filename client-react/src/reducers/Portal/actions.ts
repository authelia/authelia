import { createAction } from 'typesafe-actions';
import {
  AUTHENTICATE_REQUEST,
  AUTHENTICATE_SUCCESS,
  AUTHENTICATE_FAILURE,
  FETCH_STATE_REQUEST,
  FETCH_STATE_SUCCESS,
  FETCH_STATE_FAILURE,
  LOGOUT_REQUEST,
  LOGOUT_SUCCESS,
  LOGOUT_FAILURE
} from "../constants";
import RemoteState from './RemoteState';

/*     FETCH_STATE    */
export const fetchState = createAction(FETCH_STATE_REQUEST);
export const fetchStateSuccess = createAction(FETCH_STATE_SUCCESS, resolve => {
  return (state: RemoteState) => {
    return resolve(state);
  }
});
export const fetchStateFailure = createAction(FETCH_STATE_FAILURE, resolve => {
  return (err: string) => {
    return resolve(err);
  }
})

/*     AUTHENTICATE_REQUEST    */
export const authenticate = createAction(AUTHENTICATE_REQUEST);
export const authenticateSuccess = createAction(AUTHENTICATE_SUCCESS);
export const authenticateFailure = createAction(AUTHENTICATE_FAILURE, resolve => {
  return (error: string) => resolve(error);
});


/*     AUTHENTICATE_REQUEST    */
export const logout = createAction(LOGOUT_REQUEST);
export const logoutSuccess = createAction(LOGOUT_SUCCESS);
export const logoutFailure = createAction(LOGOUT_FAILURE, resolve => {
  return (error: string) => resolve(error);
});
