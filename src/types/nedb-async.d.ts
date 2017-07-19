import Nedb = require("nedb");
import BluebirdPromise = require("bluebird");

declare module "nedb" {
    export class NedbAsync extends Nedb {
        constructor(pathOrOptions?: string | Nedb.DataStoreOptions);
        updateAsync(query: any, updateQuery: any, options?: Nedb.UpdateOptions): BluebirdPromise<any>;
        findOneAsync<T>(query: any): BluebirdPromise<T>;
        insertAsync<T>(newDoc: T): BluebirdPromise<any>;
        removeAsync(query: any): BluebirdPromise<any>;
        countAsync(query: any): BluebirdPromise<number>
    }
}