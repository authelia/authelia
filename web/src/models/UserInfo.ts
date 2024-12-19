import { REGEX } from "@constants/Regex";
import { SecondFactorMethod } from "@models/Methods";

export interface UserInfo {
    username: string;
    display_name: string;
    emails: string[];
    groups: string[];
    disabled?: boolean;
    last_logged_in?: Date;
    password_change_required?: boolean;
    last_password_change?: Date;
    logout_required?: boolean;
    user_created_at?: Date;
    method: SecondFactorMethod;
    has_webauthn: boolean;
    has_totp: boolean;
    has_duo: boolean;
}

export function ValidateUsername(username: string) {
    if (!username) return false;

    username = username.trim();
    if (!username) return false;

    if (username.includes("@")) {
        return REGEX.EMAIL.test(username);
    }
    return REGEX.USERNAME.test(username);
}

export function ValidateDisplayName(displayName: string) {
    if (!displayName) return false;

    displayName = displayName.trim();
    if (!displayName) return false;

    return REGEX.DISPLAY_NAME.test(displayName);
}

export function ValidateEmail(email: string) {
    email = email.trim();
    if (!email) return false;

    return REGEX.EMAIL.test(email);
}

export function ValidateGroup(group: string) {
    if (!group) return false;

    group = group.trim();
    if (!group) return false;

    return REGEX.GROUP.test(group);
}
