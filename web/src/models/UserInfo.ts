import { SecondFactorMethod } from "./Methods";

export interface UserInfo {
    display_name: string;
    method: SecondFactorMethod;
    has_u2f: boolean;
    has_totp: boolean;
}
