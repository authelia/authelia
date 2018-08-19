
import MongoDB = require("mongodb");
import { IMongoClient } from "./IMongoClient";
import Bluebird = require("bluebird");
import { AUTHENTICATION_FAILED } from "../../../../../shared/UserMessages";
import { IGlobalLogger } from "../../logging/IGlobalLogger";

export class MongoClient implements IMongoClient {
  private url: string;
  private databaseName: string;

  private database: MongoDB.Db;
  private client: MongoDB.MongoClient;
  private logger: IGlobalLogger;

  constructor(
    url: string,
    databaseName: string,
    logger: IGlobalLogger) {

    this.url = url;
    this.databaseName = databaseName;
    this.logger = logger;
  }

  connect(): Bluebird<void> {
    const that = this;
    const connectAsync = Bluebird.promisify(MongoDB.MongoClient.connect);
    return connectAsync(this.url)
      .then(function (client: MongoDB.MongoClient) {
        that.database = client.db(that.databaseName);
        that.database.on("close", () => {
          that.logger.info("[MongoClient] Lost connection.");
        });
        that.database.on("reconnect", () => {
          that.logger.info("[MongoClient] Reconnected.");
        });
        that.client = client;
      });
  }

  close(): Bluebird<void> {
    if (this.client) {
      this.client.close();
      this.database = undefined;
      this.client = undefined;
    }
    return Bluebird.resolve();
  }

  collection(name: string): Bluebird<MongoDB.Collection> {
    if (!this.client) {
      const that = this;
      return this.connect()
        .then(() => Bluebird.resolve(that.database.collection(name)));
    }

    return Bluebird.resolve(this.database.collection(name));
  }
}