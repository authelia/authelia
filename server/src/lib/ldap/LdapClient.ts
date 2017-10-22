import LdapJs = require("ldapjs");
import EventEmitter = require("events");
import BluebirdPromise = require("bluebird");
import { ILdapClient } from "./ILdapClient";
import Exceptions = require("../Exceptions");

declare module "ldapjs" {
  export interface ClientAsync {
    on(event: string, callback: (data?: any) => void): void;
    bindAsync(username: string, password: string): BluebirdPromise<void>;
    unbindAsync(): BluebirdPromise<void>;
    searchAsync(base: string, query: LdapJs.SearchOptions): BluebirdPromise<EventEmitter>;
    modifyAsync(userdn: string, change: LdapJs.Change): BluebirdPromise<void>;
  }
}

interface SearchEntry {
  object: any;
}

export class LdapClient implements ILdapClient {
  private client: LdapJs.ClientAsync;

  constructor(url: string, ldapjs: typeof LdapJs) {
    const ldapClient = ldapjs.createClient({
      url: url,
      reconnect: true
    });

    /*const clientLogger = (ldapClient as any).log;
    if (clientLogger) {
      clientLogger.level("trace");
    }*/

    this.client = BluebirdPromise.promisifyAll(ldapClient) as any;
  }

  bindAsync(username: string, password: string): BluebirdPromise<void> {
    return this.client.bindAsync(username, password);
  }

  unbindAsync(): BluebirdPromise<void> {
    return this.client.unbindAsync();
  }

  searchAsync(base: string, query: any): BluebirdPromise<any[]> {
    const that = this;
    return this.client.searchAsync(base, query)
      .then(function (res: EventEmitter) {
        const doc: SearchEntry[] = [];
        return new BluebirdPromise<any[]>((resolve, reject) => {
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
        return BluebirdPromise.reject(new Exceptions.LdapSearchError(err.message));
      });
  }

  modifyAsync(dn: string, changeRequest: any): BluebirdPromise<void> {
    return this.client.modifyAsync(dn, changeRequest);
  }
}