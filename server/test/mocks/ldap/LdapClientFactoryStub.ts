import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import { ILdapClientFactory } from "../../../src/lib/ldap/ILdapClientFactory";
import { ILdapClient } from "../../../src/lib/ldap/ILdapClient";

export class LdapClientFactoryStub implements ILdapClientFactory {
  createStub: Sinon.SinonStub;

  constructor() {
    this.createStub = Sinon.stub();
  }

  create(): ILdapClient {
    return this.createStub();
  }
}