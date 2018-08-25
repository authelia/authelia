import Sinon = require("sinon");

import { ISession } from "./ISession";
import { ISessionFactory } from "./ISessionFactory";

export class SessionFactoryStub implements ISessionFactory {
  createStub: Sinon.SinonStub;

  constructor() {
    this.createStub = Sinon.stub();
  }

  create(userDN: string, password: string): ISession {
    return this.createStub(userDN, password);
  }
}