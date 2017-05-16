import Nedb = require("nedb");
import * as Promise from "bluebird";

declare module "nedb" {
    export class NedbAsync extends Nedb {
        constructor(pathOrOptions?: string | Nedb.DataStoreOptions);
        updateAsync(query: any, updateQuery: any, options?: Nedb.UpdateOptions): Promise<any>;
        findOneAsync(query: any): Promise<any>;
        insertAsync<T>(newDoc: T): Promise<any>;
        removeAsync(query: any): Promise<any>;
    }
}