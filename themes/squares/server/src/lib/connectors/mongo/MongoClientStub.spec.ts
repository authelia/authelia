import Sinon = require("sinon");
import MongoDB = require("mongodb");
import Bluebird = require("bluebird");
import { IMongoClient } from "../../../../src/lib/connectors/mongo/IMongoClient";

export class MongoClientStub implements IMongoClient {
  public collectionStub: Sinon.SinonStub;

  constructor() {
    this.collectionStub = Sinon.stub();
  }

  collection(name: string): Bluebird<MongoDB.Collection> {
    return this.collectionStub(name);
  }
}