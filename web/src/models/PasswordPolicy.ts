export enum PasswordPolicyMode {
    Disabled = 0,
    Standard = 1,
    ZXCVBN = 2,
}

export interface PasswordPolicyConfiguration {
    mode: PasswordPolicyMode;
    min_length: number;
    max_length: number;
    require_uppercase: boolean;
    require_lowercase: boolean;
    require_number: boolean;
    require_special: boolean;
}
