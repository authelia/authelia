/* eslint-disable no-unused-vars */
export enum PasswordPolicyMode {
    Disabled = 0,
    Standard = 1,
    ZXCVBN = 2,
}

export interface PasswordPolicyConfiguration {
    mode: PasswordPolicyMode;
    min_length: number;
    max_length: number;
    min_score: number;
    require_uppercase: boolean;
    require_lowercase: boolean;
    require_number: boolean;
    require_special: boolean;
}
