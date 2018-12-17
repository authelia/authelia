import BluebirdPromise = require("bluebird");
import { ICollection } from "../ICollection";
import { ICollectionFactory } from "../ICollectionFactory";
import { MongoCollection } from "./MongoCollection";
import path = require("path");
import MongoDB = require("mongodb");
import { IMongoClient } from "../../connectors/mongo/IMongoClient";

export class MongoCollectionFactory implements ICollectionFactory {
  private mongoClient: IMongoClient;

  constructor(mongoClient: IMongoClient) {
    this.mongoClient = mongoClient;
  }

  build(collectionName: string): ICollection {
    return new MongoCollection(collectionName, this.mongoClient);
  }
}