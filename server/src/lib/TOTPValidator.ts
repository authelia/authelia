import Speakeasy = require("speakeasy");
import BluebirdPromise = require("bluebird");

const TOTP_ENCODING = "base32";
const WINDOW: number = 1;

export class TOTPValidator {
  private speakeasy: typeof Speakeasy;

  constructor(speakeasy: typeof Speakeasy) {
    this.speakeasy = speakeasy;
  }

  validate(token: string, secret: string): BluebirdPromise<void> {
    const isValid = this.speakeasy.totp.verify({
      secret: secret,
      encoding: TOTP_ENCODING,
      token: token,
      window: WINDOW
    } as any);

    if (isValid)
      return BluebirdPromise.resolve();
    else
      return BluebirdPromise.reject(new Error("Wrong TOTP token."));
  }
}