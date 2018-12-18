import { TOTPSecret } from "../../../../types/TOTPSecret";

export interface ITotpHandler {
  generate(label: string, issuer: string): TOTPSecret;
  validate(token: string, secret: string): boolean;
}