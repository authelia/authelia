
import MongoDB = require("mongodb");
import BluebirdPromise = require("bluebird");
import { IMongoClient } from "./IMongoClient";
import { IMongoConnector } from "./IMongoConnector";
import { MongoClient } from "./MongoClient";

export class MongoConnector implements IMongoConnector {
  private url: string;
  private client: MongoDB.MongoClient;

  constructor(url: string) {
    this.url = url;
  }

  connect(databaseName: string): BluebirdPromise<IMongoClient> {
    const that = this;
    const connectAsync = BluebirdPromise.promisify(MongoDB.MongoClient.connect);
    return connectAsync(this.url)
      .then(function (client: MongoDB.MongoClient) {
        that.client = client;
        const db = client.db(databaseName);
        return BluebirdPromise.resolve(new MongoClient(db));
      });
  }

  close(): BluebirdPromise<void> {
    this.client.close();
    return BluebirdPromise.resolve();
  }
}