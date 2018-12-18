
export interface TOTPSecret {
    ascii: string;
    hex: string;
    base32: string;
    qr_code_ascii: string;
    qr_code_hex: string;
    qr_code_base32: string;
    google_auth_qr: string;
    otpauth_url: string;
  }