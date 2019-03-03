import { createAction } from "typesafe-actions";
import { GENERATE_TOTP_SECRET_REQUEST, GENERATE_TOTP_SECRET_SUCCESS, GENERATE_TOTP_SECRET_FAILURE } from "../../constants";
import { Secret } from "../../../views/OneTimePasswordRegistrationView/Secret";


/*     GENERATE_TOTP_SECRET_REQUEST    */
export const generateTotpSecret = createAction(GENERATE_TOTP_SECRET_REQUEST);
export const generateTotpSecretSuccess = createAction(GENERATE_TOTP_SECRET_SUCCESS, resolve => {
  return (secret: Secret) => resolve(secret);
});
export const generateTotpSecretFailure = createAction(GENERATE_TOTP_SECRET_FAILURE, resolve => {
  return (error: string) => resolve(error);
});
