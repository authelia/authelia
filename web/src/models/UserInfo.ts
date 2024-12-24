import { SecondFactorMethod } from "@models/Methods";

export interface UserInfo {
    display_name: string;
    emails: string[];
    method: SecondFactorMethod;
    has_webauthn: boolean;
    has_totp: boolean;
    has_duo: boolean;
}
