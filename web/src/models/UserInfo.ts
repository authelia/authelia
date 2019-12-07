import { SecondFactorMethod } from "./Methods";

export interface UserInfo {
    method: SecondFactorMethod;
    has_u2f: boolean;
    has_totp: boolean;
}
