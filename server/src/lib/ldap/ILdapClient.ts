
import BluebirdPromise = require("bluebird");
import EventEmitter = require("events");

export interface ILdapClient {
  bindAsync(username: string, password: string): BluebirdPromise<void>;
  unbindAsync(): BluebirdPromise<void>;
  searchAsync(base: string, query: any): BluebirdPromise<any[]>;
  modifyAsync(dn: string, changeRequest: any): BluebirdPromise<void>;
}