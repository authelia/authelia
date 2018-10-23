import Bluebird = require("bluebird");
import { ICollection } from "../ICollection";
import MongoDB = require("mongodb");
import { IMongoClient } from "../../connectors/mongo/IMongoClient";


export class MongoCollection implements ICollection {
  private mongoClient: IMongoClient;
  private collectionName: string;

  constructor(collectionName: string, mongoClient: IMongoClient) {
    this.collectionName = collectionName;
    this.mongoClient = mongoClient;
  }

  private collection(): Bluebird<MongoDB.Collection> {
    return this.mongoClient.collection(this.collectionName);
  }

  find(query: any, sortKeys?: any, count?: number): Bluebird<any> {
    return this.collection()
      .then((collection) => collection.find(query).sort(sortKeys).limit(count))
      .then((query) => query.toArray());
  }

  findOne(query: any): Bluebird<any> {
    return this.collection()
      .then((collection) => collection.findOne(query));
  }

  update(query: any, updateQuery: any, options?: any): Bluebird<any> {
    return this.collection()
      .then((collection) => collection.update(query, updateQuery, options));
  }

  remove(query: any): Bluebird<any> {
    return this.collection()
      .then((collection) => collection.remove(query));
  }

  insert(document: any): Bluebird<any> {
    return this.collection()
      .then((collection) => collection.insertOne(document));
  }

  count(query: any): Bluebird<any> {
    return this.collection()
      .then((collection) => collection.count(query));
  }
}