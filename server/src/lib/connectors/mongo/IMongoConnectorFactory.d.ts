import { IMongoConnector } from "./IMongoConnector";

export interface IMongoConnectorFactory {
  create(url: string): IMongoConnector;
}