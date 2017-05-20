import ldapjs = require("ldapjs");
import * as Promise from "bluebird";
import { EventEmitter } from "events";

declare module "ldapjs" {
    export interface ClientAsync {
        bindAsync(username: string, password: string): Promise<void>;
        searchAsync(base: string, query: ldapjs.SearchOptions): Promise<EventEmitter>;
        modifyAsync(userdn: string, change: ldapjs.Change): Promise<void>;
    }
}