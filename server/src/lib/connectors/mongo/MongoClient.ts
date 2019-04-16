
import MongoDB = require("mongodb");
import { IMongoClient } from "./IMongoClient";
import Bluebird = require("bluebird");
import { IGlobalLogger } from "../../logging/IGlobalLogger";
import { MongoStorageConfiguration } from "../../configuration/schema/StorageConfiguration";

export class MongoClient implements IMongoClient {
  private configuration: MongoStorageConfiguration;

  private database: MongoDB.Db;
  private client: MongoDB.MongoClient;
  private logger: IGlobalLogger;

  constructor(
    configuration: MongoStorageConfiguration,
    logger: IGlobalLogger) {

    this.configuration = configuration;
    this.logger = logger;
  }

  connect(): Bluebird<void> {
    const that = this;
    const options: MongoDB.MongoClientOptions = {};
    if (that.configuration.auth) {
      options["auth"] = {
        user: that.configuration.auth.username,
        password: that.configuration.auth.password
      };
    }

    return new Bluebird((resolve, reject) => {
        MongoDB.MongoClient.connect(
          this.configuration.url,
          options,
          function(err, client) {
          if (err) {
            reject(err);
            return;
          }
          resolve(client);
        });
      })
      .then(function (client: MongoDB.MongoClient) {
        that.database = client.db(that.configuration.database);
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