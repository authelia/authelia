
import MongoDB = require("mongodb");
import { IMongoClient } from "./IMongoClient";

export class MongoClient implements IMongoClient {
  private db: MongoDB.Db;

  constructor(db: MongoDB.Db) {
    this.db = db;
  }

  collection(name: string): MongoDB.Collection {
    return this.db.collection(name);
  }
}