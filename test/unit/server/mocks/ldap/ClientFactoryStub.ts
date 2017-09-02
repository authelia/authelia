
import { IClient } from "../../../../../src/server/lib/ldap/IClient";
import { IClientFactory } from "../../../../../src/server/lib/ldap/IClientFactory";
import Sinon = require("sinon");

export class ClientFactoryStub implements IClientFactory {
  createStub: Sinon.SinonStub;

  constructor() {
    this.createStub = Sinon.stub();
  }

  create(userDN: string, password: string): IClient {
    return this.createStub(userDN, password);
  }
}