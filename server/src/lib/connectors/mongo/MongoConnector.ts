
import MongoDB = require("mongodb");
import BluebirdPromise = require("bluebird");
import { IMongoClient } from "./IMongoClient";
import { IMongoConnector } from "./IMongoConnector";
import { MongoClient } from "./MongoClient";

export class MongoConnector implements IMongoConnector {
  private url: string;

  constructor(url: string) {
    this.url = url;
  }

  connect(): BluebirdPromise<IMongoClient> {
    const connectAsync = BluebirdPromise.promisify(MongoDB.MongoClient.connect);
    return connectAsync(this.url)
      .then(function (db: MongoDB.Db) {
        return BluebirdPromise.resolve(new MongoClient(db));
      });
  }
}