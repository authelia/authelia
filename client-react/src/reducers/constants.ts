
export const FETCH_STATE_REQUEST = '@portal/fetch_state_request';
export const FETCH_STATE_SUCCESS = '@portal/fetch_state_success';
export const FETCH_STATE_FAILURE = '@portal/fetch_state_failure';

export const AUTHENTICATE_REQUEST = '@portal/authenticate_request';
export const AUTHENTICATE_SUCCESS = '@portal/authenticate_success';
export const AUTHENTICATE_FAILURE = '@portal/authenticate_failure';

// SECOND FACTOR PAGE
export const SET_SECURITY_KEY_SUPPORTED = '@portal/second_factor/set_security_key_supported';

export const SECURITY_KEY_SIGN = '@portal/second_factor/security_key_sign';
export const SECURITY_KEY_SIGN_SUCCESS = '@portal/second_factor/security_key_sign_success';
export const SECURITY_KEY_SIGN_FAILURE = '@portal/second_factor/security_key_sign_failure';

export const LOGOUT_REQUEST = '@portal/logout_request';
export const LOGOUT_SUCCESS = '@portal/logout_success';
export const LOGOUT_FAILURE = '@portal/logout_failure';

// TOTP REGISTRATION
export const GENERATE_TOTP_SECRET_REQUEST = '@portal/generate_totp_secret_request';
export const GENERATE_TOTP_SECRET_SUCCESS = '@portal/generate_totp_secret_success';
export const GENERATE_TOTP_SECRET_FAILURE = '@portal/generate_totp_secret_failure';

// U2F REGISTRATION
export const REGISTER_SECURITY_KEY_REQUEST = '@portal/security_key_registration/register_request';
export const REGISTER_SECURITY_KEY_SUCCESS = '@portal/security_key_registration/register_success';
export const REGISTER_SECURITY_KEY_FAILURE = '@portal/security_key_registration/register_failed';