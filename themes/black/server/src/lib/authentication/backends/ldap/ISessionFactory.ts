
import { ISession } from "./ISession";

export interface ISessionFactory {
  create(userDN: string, password: string): ISession;
}