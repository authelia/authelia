
export const FETCH_STATE_REQUEST = '@portal/fetch_state_request';
export const FETCH_STATE_SUCCESS = '@portal/fetch_state_success';
export const FETCH_STATE_FAILURE = '@portal/fetch_state_failure';

// AUTHENTICATION PROCESS
export const FIRST_FACTOR_SET_USERNAME = "@portal/first_factor/set_username";
export const FIRST_FACTOR_SET_PASSWORD = "@portal/first_factor/set_password";

export const AUTHENTICATE_REQUEST = '@portal/first_factor/authenticate_request';
export const AUTHENTICATE_SUCCESS = '@portal/first_factor/authenticate_success';
export const AUTHENTICATE_FAILURE = '@portal/first_factor/authenticate_failure';

// SECOND FACTOR PAGE
export const SET_SECURITY_KEY_SUPPORTED = '@portal/second_factor/set_security_key_supported';
export const SET_USE_ANOTHER_METHOD = '@portal/second_factor/set_use_another_method';

export const GET_AVAILABLE_METHODS = '@portal/second_factor/get_available_methods';
export const GET_AVAILABLE_METHODS_SUCCESS = '@portal/second_factor/get_available_methods_success';
export const GET_AVAILABLE_METHODS_FAILURE = '@portal/second_factor/get_available_methods_failure';

export const GET_PREFERED_METHOD = '@portal/second_factor/get_prefered_method';
export const GET_PREFERED_METHOD_SUCCESS = '@portal/second_factor/get_prefered_method_success';
export const GET_PREFERED_METHOD_FAILURE = '@portal/second_factor/get_prefered_method_failure';

export const SET_PREFERED_METHOD = '@portal/second_factor/set_prefered_method';
export const SET_PREFERED_METHOD_SUCCESS = '@portal/second_factor/set_prefered_method_success';
export const SET_PREFERED_METHOD_FAILURE = '@portal/second_factor/set_prefered_method_failure';

export const SECURITY_KEY_SIGN = '@portal/second_factor/security_key_sign';
export const SECURITY_KEY_SIGN_SUCCESS = '@portal/second_factor/security_key_sign_success';
export const SECURITY_KEY_SIGN_FAILURE = '@portal/second_factor/security_key_sign_failure';

export const ONE_TIME_PASSWORD_VERIFICATION_REQUEST = '@portal/second_factor/one_time_password_verification_request';
export const ONE_TIME_PASSWORD_VERIFICATION_SUCCESS = '@portal/second_factor/one_time_password_verification_success';
export const ONE_TIME_PASSWORD_VERIFICATION_FAILURE = '@portal/second_factor/one_time_password_verification_failure';

export const TRIGGER_DUO_PUSH_AUTH = '@portal/second_factor/trigger_duo_push_auth_request';
export const TRIGGER_DUO_PUSH_AUTH_SUCCESS = '@portal/second_factor/trigger_duo_push_auth_request_success';
export const TRIGGER_DUO_PUSH_AUTH_FAILURE = '@portal/second_factor/trigger_duo_push_auth_request_failure';

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

// FORGOT PASSWORD
export const FORGOT_PASSWORD_REQUEST = '@portal/forgot_password/forgot_password_request';
export const FORGOT_PASSWORD_SUCCESS = '@portal/forgot_password/forgot_password_success';
export const FORGOT_PASSWORD_FAILURE = '@portal/forgot_password/forgot_password_failure';

// FORGOT PASSWORD
export const RESET_PASSWORD_REQUEST = '@portal/forgot_password/reset_password_request';
export const RESET_PASSWORD_SUCCESS = '@portal/forgot_password/reset_password_success';
export const RESET_PASSWORD_FAILURE = '@portal/forgot_password/reset_password_failure';