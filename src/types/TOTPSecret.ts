
export interface TOTPSecret {
    base32: string;
    ascii: string;
    otpauth_url?: string;
}