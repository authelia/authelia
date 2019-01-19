import { createAction } from 'typesafe-actions';
import {
  AUTHENTICATE_REQUEST,
  AUTHENTICATE_SUCCESS,
  AUTHENTICATE_FAILURE
} from "../../constants";

/*     AUTHENTICATE_REQUEST    */
export const authenticate = createAction(AUTHENTICATE_REQUEST);
export const authenticateSuccess = createAction(AUTHENTICATE_SUCCESS);
export const authenticateFailure = createAction(AUTHENTICATE_FAILURE, resolve => {
  return (error: string) => resolve(error);
});
