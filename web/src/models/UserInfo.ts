import { REGEX } from "@constants/Regex";
import { SecondFactorMethod } from "@models/Methods";

export interface UserInfo {
    username: string;
    display_name: string;
    emails: string[];
    groups: string[];
    disabled?: boolean;
    last_logged_in?: Date;
    last_password_change?: Date;
    user_created_at?: Date;
    method: SecondFactorMethod;
    has_webauthn: boolean;
    has_totp: boolean;
    has_duo: boolean;
}

export function ValidateUsername(username: string) {
    if (!username) return false;

    if (!username.trim()) return false;

    if (username.includes("@")) {
        return REGEX.EMAIL.test(username);
    }
    return REGEX.USERNAME.test(username);
}

export function ValidateDisplayName(displayName: string) {
    if (!displayName) return false;

    if (!displayName.trim()) return false;

    return REGEX.DISPLAY_NAME.test(displayName);
}

export function ValidateEmail(email: string) {
    if (!email.trim()) return false;

    return REGEX.EMAIL.test(email);
}

export function ValidateGroup(group: string) {
    if (!group) return false;

    if (!group.trim()) return false;

    return REGEX.GROUP.test(group);
}
