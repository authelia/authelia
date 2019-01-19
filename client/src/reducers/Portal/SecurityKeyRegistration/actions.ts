import { createAction } from "typesafe-actions";
import { REGISTER_SECURITY_KEY_REQUEST, REGISTER_SECURITY_KEY_SUCCESS, REGISTER_SECURITY_KEY_FAILURE } from "../../constants";

export const registerSecurityKey = createAction(REGISTER_SECURITY_KEY_REQUEST);
export const registerSecurityKeySuccess = createAction(REGISTER_SECURITY_KEY_SUCCESS);
export const registerSecurityKeyFailure = createAction(REGISTER_SECURITY_KEY_FAILURE, resolve => {
  return (error: string) => resolve(error);
});
