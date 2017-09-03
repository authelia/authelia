
import { IClient } from "./IClient";

export interface IClientFactory {
  create(userDN: string, password: string): IClient;
}