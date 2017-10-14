
import Speakeasy = require("speakeasy");
import BluebirdPromise = require("bluebird");
import { TOTPSecret } from "../../types/TOTPSecret";

interface GenerateSecretOptions {
  length?: number;
  symbols?: boolean;
  otpauth_url?: boolean;
  name?: string;
  issuer?: string;
}

export class TOTPGenerator {
  private speakeasy: typeof Speakeasy;

  constructor(speakeasy: typeof Speakeasy) {
    this.speakeasy = speakeasy;
  }

  generate(options?: GenerateSecretOptions): TOTPSecret {
    return this.speakeasy.generateSecret(options);
  }
}