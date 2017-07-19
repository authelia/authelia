
import BluebirdPromise = require("bluebird");
import { IMongoConnectorFactory } from "./IMongoConnectorFactory";
import { IMongoConnector } from "./IMongoConnector";
import { MongoConnector } from "./MongoConnector";
import MongoDB = require("mongodb");

export class MongoConnectorFactory implements IMongoConnectorFactory {
  create(url: string): IMongoConnector {
    return new MongoConnector(url);
  }
}