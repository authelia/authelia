import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import { ITotpHandler, GenerateSecretOptions } from "../../src/lib/authentication/totp/ITotpHandler";
import { TOTPSecret } from "../../types/TOTPSecret";

export class TotpHandlerStub implements ITotpHandler {
  generateStub: Sinon.SinonStub;
  validateStub: Sinon.SinonStub;

  constructor() {
    this.generateStub = Sinon.stub();
    this.validateStub = Sinon.stub();
  }

  generate(options?: GenerateSecretOptions): TOTPSecret {
    return this.generateStub(options);
  }

  validate(token: string, secret: string): boolean {
    return this.validateStub(token, secret);
  }
}