import { TOTPSecret } from "../../../../types/TOTPSecret";

export interface GenerateSecretOptions {
  length?: number;
  symbols?: boolean;
  otpauth_url?: boolean;
  name?: string;
  issuer?: string;
}

export interface ITotpHandler {
  generate(options?: GenerateSecretOptions): TOTPSecret;
  validate(token: string, secret: string): boolean;
}