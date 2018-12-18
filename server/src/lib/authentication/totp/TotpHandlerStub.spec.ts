import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import { ITotpHandler } from "./ITotpHandler";
import { TOTPSecret } from "../../../../types/TOTPSecret";

export class TotpHandlerStub implements ITotpHandler {
  generateStub: Sinon.SinonStub;
  validateStub: Sinon.SinonStub;

  constructor() {
    this.generateStub = Sinon.stub();
    this.validateStub = Sinon.stub();
  }

  generate(label: string, issuer: string): TOTPSecret {
    return this.generateStub(label, issuer);
  }

  validate(token: string, secret: string): boolean {
    return this.validateStub(token, secret);
  }
}