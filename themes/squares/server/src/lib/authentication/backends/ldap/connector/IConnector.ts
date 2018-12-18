import Bluebird = require("bluebird");
import EventEmitter = require("events");

export interface IConnector {
  bindAsync(username: string, password: string): Bluebird<void>;
  unbindAsync(): Bluebird<void>;
  searchAsync(base: string, query: any): Bluebird<any[]>;
  modifyAsync(dn: string, changeRequest: any): Bluebird<void>;
}