import Sinon = require("sinon");
import MongoDB = require("mongodb");
import { IMongoClient } from "../../../../../../src/server/lib/connectors/mongo/IMongoClient";

export class MongoClientStub implements IMongoClient {
  public collectionStub: Sinon.SinonStub;

  constructor() {
    this.collectionStub = Sinon.stub();
  }

  collection(name: string): MongoDB.Collection {
    return this.collectionStub(name);
  }
}