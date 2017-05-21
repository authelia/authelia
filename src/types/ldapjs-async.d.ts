import ldapjs = require("ldapjs");
import * as BluebirdPromise from "bluebird";
import { EventEmitter } from "events";

declare module "ldapjs" {
    export interface ClientAsync {
        bindAsync(username: string, password: string): BluebirdPromise<void>;
        searchAsync(base: string, query: ldapjs.SearchOptions): BluebirdPromise<EventEmitter>;
        modifyAsync(userdn: string, change: ldapjs.Change): BluebirdPromise<void>;
    }
}