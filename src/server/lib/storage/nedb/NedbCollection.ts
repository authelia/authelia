import BluebirdPromise = require("bluebird");
import { ICollection } from "../ICollection";
import Nedb = require("nedb");
import { NedbAsync } from "nedb";


export class NedbCollection implements ICollection {
  private collection: NedbAsync;

  constructor(options: Nedb.DataStoreOptions) {
    this.collection = BluebirdPromise.promisifyAll(new Nedb(options)) as NedbAsync;
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