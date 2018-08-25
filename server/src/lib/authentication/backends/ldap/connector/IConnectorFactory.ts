import { IConnector } from "./IConnector";

export interface IConnectorFactory {
  create(): IConnector;
}