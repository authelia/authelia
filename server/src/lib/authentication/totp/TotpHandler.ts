import { ITotpHandler, GenerateSecretOptions } from "./ITotpHandler";
import { TOTPSecret } from "../../../../types/TOTPSecret";
import Speakeasy = require("speakeasy");

const TOTP_ENCODING = "base32";
const WINDOW: number = 1;

export class TotpHandler implements ITotpHandler {
  private speakeasy: typeof Speakeasy;

  constructor(speakeasy: typeof Speakeasy) {
    this.speakeasy = speakeasy;
  }

  generate(options?: GenerateSecretOptions): TOTPSecret {
    return this.speakeasy.generateSecret(options);
  }

  validate(token: string, secret: string): boolean {
    return this.speakeasy.totp.verify({
      secret: secret,
      encoding: TOTP_ENCODING,
      token: token,
      window: WINDOW
    } as any);
  }
}