import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import { IRegulator } from "../../src/lib/regulation/IRegulator";

export class RegulatorStub implements IRegulator {
  markStub: Sinon.SinonStub;
  regulateStub: Sinon.SinonStub;

  constructor() {
    this.markStub = Sinon.stub();
    this.regulateStub = Sinon.stub();
  }

  mark(userId: string, isAuthenticationSuccessful: boolean): BluebirdPromise<void> {
    return this.markStub(userId, isAuthenticationSuccessful);
  }

  regulate(userId: string): BluebirdPromise<void> {
    return this.regulateStub(userId);
  }
}