import { ICollectionFactory } from "./ICollectionFactory";
import { NedbCollectionFactory } from "./nedb/NedbCollectionFactory";
import { MongoCollectionFactory } from "./mongo/MongoCollectionFactory";
import { IMongoClient } from "../connectors/mongo/IMongoClient";


export class CollectionFactoryFactory {
  static createNedb(options: Nedb.DataStoreOptions): ICollectionFactory {
    return new NedbCollectionFactory(options);
  }

  static createMongo(client: IMongoClient): ICollectionFactory {
    return new MongoCollectionFactory(client);
  }
}