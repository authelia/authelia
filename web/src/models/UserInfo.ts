import { SecondFactorMethod } from "@models/Methods";

export interface UserInfo {
    display_name: string;
    emails: string[];
    disabled?: boolean;
    last_logged_in?: boolean;
    password_change_required?: boolean;
    last_password_change?: boolean;
    logout_required?: boolean;
    method: SecondFactorMethod;
    has_webauthn: boolean;
    has_totp: boolean;
    has_duo: boolean;
}
