import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import { ILdapClientFactory } from "./ILdapClientFactory";
import { ILdapClient } from "./ILdapClient";

export class LdapClientFactoryStub implements ILdapClientFactory {
  createStub: Sinon.SinonStub;

  constructor() {
    this.createStub = Sinon.stub();
  }

  create(): ILdapClient {
    return this.createStub();
  }
}