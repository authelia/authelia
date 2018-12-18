import BluebirdPromise = require("bluebird");
import Sinon = require("sinon");

import { IConnector } from "./IConnector";

export class ConnectorStub implements IConnector {
  bindAsyncStub: Sinon.SinonStub;
  unbindAsyncStub: Sinon.SinonStub;
  searchAsyncStub: Sinon.SinonStub;
  modifyAsyncStub: Sinon.SinonStub;

  constructor() {
    this.bindAsyncStub = Sinon.stub();
    this.unbindAsyncStub = Sinon.stub();
    this.searchAsyncStub = Sinon.stub();
    this.modifyAsyncStub = Sinon.stub();
  }

  bindAsync(username: string, password: string): BluebirdPromise<void> {
    return this.bindAsyncStub(username, password);
  }

  unbindAsync(): BluebirdPromise<void> {
    return this.unbindAsyncStub();
  }

  searchAsync(base: string, query: any): BluebirdPromise<any[]> {
    return this.searchAsyncStub(base, query);
  }

  modifyAsync(dn: string, changeRequest: any): BluebirdPromise<void> {
    return this.modifyAsyncStub(dn, changeRequest);
  }
}