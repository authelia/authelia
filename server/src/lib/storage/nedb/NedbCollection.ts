import BluebirdPromise = require("bluebird");
import { ICollection } from "../ICollection";
import Nedb = require("nedb");

declare module "nedb" {
    export class NedbAsync extends Nedb {
        constructor(pathOrOptions?: string | Nedb.DataStoreOptions);
        updateAsync(query: any, updateQuery: any, options?: Nedb.UpdateOptions): BluebirdPromise<any>;
        findOneAsync<T>(query: any): BluebirdPromise<T>;
        insertAsync<T>(newDoc: T): BluebirdPromise<any>;
        removeAsync(query: any): BluebirdPromise<any>;
        countAsync(query: any): BluebirdPromise<number>;
    }
}

export class NedbCollection implements ICollection {
  private collection: Nedb.NedbAsync;

  constructor(options: Nedb.DataStoreOptions) {
    this.collection = BluebirdPromise.promisifyAll(new Nedb(options)) as Nedb.NedbAsync;
  }

  find(query: any, sortKeys?: any, count?: number): BluebirdPromise<any> {
    const q = this.collection.find(query).sort(sortKeys).limit(count);
    return BluebirdPromise.promisify(q.exec, { context: q })();
  }

  findOne(query: any): BluebirdPromise<any> {
    return this.collection.findOneAsync(query);
  }

  update(query: any, updateQuery: any, options?: any): BluebirdPromise<any> {
    return this.collection.updateAsync(query, updateQuery, options);
  }

  remove(query: any): BluebirdPromise<any> {
    return this.collection.removeAsync(query);
  }

  insert(document: any): BluebirdPromise<any> {
    return this.collection.insertAsync(document);
  }

  count(query: any): BluebirdPromise<number> {
    return this.collection.countAsync(query);
  }
}