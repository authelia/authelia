import LdapJs = require("ldapjs");
import EventEmitter = require("events");
import Bluebird = require("bluebird");
import { IConnector } from "./IConnector";
import Exceptions = require("../../../../Exceptions");
import { Client, ClientOptions } from "ldapjs";

interface SearchEntry {
  object: any;
}

export interface ClientAsync {
  on(event: string, callback: (data?: any) => void): void;
  bindAsync(username: string, password: string): Bluebird<void>;
  unbindAsync(): Bluebird<void>;
  searchAsync(base: string, query: LdapJs.SearchOptions): Bluebird<EventEmitter>;
  modifyAsync(userdn: string, change: LdapJs.Change): Bluebird<void>;
}

export class Connector implements IConnector {
  private client: ClientAsync;

  constructor(clientOptions: ClientOptions, ldapjs: typeof LdapJs) {
    const ldapClient: Client = ldapjs.createClient(clientOptions);

    /*const clientLogger = (ldapClient as any).log;
    if (clientLogger) {
      clientLogger.level("trace");
    }*/

    this.client = Bluebird.promisifyAll(ldapClient) as any;
  }

  bindAsync(username: string, password: string): Bluebird<void> {
    return this.client.bindAsync(username, password);
  }

  unbindAsync(): Bluebird<void> {
    return this.client.unbindAsync();
  }

  searchAsync(base: string, query: any): Bluebird<any[]> {
    const that = this;
    return this.client.searchAsync(base, query)
      .then(function (res: EventEmitter) {
        const doc: SearchEntry[] = [];
        return new Bluebird<any[]>((resolve, reject) => {
          res.on("searchEntry", function (entry: SearchEntry) {
            doc.push(entry.object);
          });
          res.on("error", function (err: Error) {
            reject(new Exceptions.LdapSearchError(err.message));
          });
          res.on("end", function () {
            resolve(doc);
          });
        });
      })
      .catch(function (err: Error) {
        return Bluebird.reject(new Exceptions.LdapSearchError(err.message));
      });
  }

  modifyAsync(dn: string, changeRequest: any): Bluebird<void> {
    return this.client.modifyAsync(dn, changeRequest);
  }
}