import BluebirdPromise = require("bluebird");
import Sinon = require("sinon");

import { IConnectorFactory } from "./IConnectorFactory";
import { IConnector } from "./IConnector";

export class ConnectorFactoryStub implements IConnectorFactory {
  createStub: Sinon.SinonStub;

  constructor() {
    this.createStub = Sinon.stub();
  }

  create(): IConnector {
    return this.createStub();
  }
}