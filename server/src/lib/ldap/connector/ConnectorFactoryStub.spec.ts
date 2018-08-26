import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");
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