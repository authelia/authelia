
import { Speakeasy } from "../../types/Dependencies";
import BluebirdPromise = require("bluebird");

const TOTP_ENCODING = "base32";

export class TOTPValidator {
  private speakeasy: Speakeasy;

  constructor(speakeasy: Speakeasy) {
    this.speakeasy = speakeasy;
  }

  validate(token: string, secret: string): BluebirdPromise<void> {
    const real_token = this.speakeasy.totp({
      secret: secret,
      encoding: TOTP_ENCODING
    });

    if (token == real_token) return BluebirdPromise.resolve();
    return BluebirdPromise.reject(new Error("Wrong challenge"));
  }
}