import { ITotpHandler } from "./ITotpHandler";
import { TOTPSecret } from "../../../../types/TOTPSecret";
import Speakeasy = require("speakeasy");

const TOTP_ENCODING = "base32";
const WINDOW: number = 1;

export class TotpHandler implements ITotpHandler {
  private speakeasy: typeof Speakeasy;

  constructor(speakeasy: typeof Speakeasy) {
    this.speakeasy = speakeasy;
  }

  generate(label: string, issuer: string): TOTPSecret {
    const secret = this.speakeasy.generateSecret({
      otpauth_url: false
    }) as TOTPSecret;

    secret.otpauth_url = this.speakeasy.otpauthURL({
      secret: secret.ascii,
      label: label,
      issuer: issuer
    });
    return secret;
  }

  validate(token: string, secret: string): boolean {
    return this.speakeasy.totp.verify({
      secret: secret,
      encoding: TOTP_ENCODING,
      token: token,
      window: WINDOW
    });
  }
}
