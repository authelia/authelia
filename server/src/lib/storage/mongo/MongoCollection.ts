
import BluebirdPromise = require("bluebird");
import { ICollection } from "../ICollection";
import MongoDB = require("mongodb");


export class MongoCollection implements ICollection {
  private collection: MongoDB.Collection;

  constructor(collection: MongoDB.Collection) {
    this.collection = collection;
  }

  find(query: any, sortKeys?: any, count?: number): BluebirdPromise<any> {
    const q = this.collection.find(query).sort(sortKeys).limit(count);
    const toArrayAsync = BluebirdPromise.promisify(q.toArray, { context: q });
    return toArrayAsync();
  }

  findOne(query: any): BluebirdPromise<any> {
    const findOneAsync = BluebirdPromise.promisify<any, any>(this.collection.findOne, { context: this.collection });
    return findOneAsync(query);
  }

  update(query: any, updateQuery: any, options?: any): BluebirdPromise<any> {
    const updateAsync = BluebirdPromise.promisify<any, any, any, any>(this.collection.update, { context: this.collection });
    return updateAsync(query, updateQuery, options);
  }

  remove(query: any): BluebirdPromise<any> {
    const removeAsync = BluebirdPromise.promisify<any, any>(this.collection.remove, { context: this.collection });
    return removeAsync(query);
  }

  insert(document: any): BluebirdPromise<any> {
    const insertAsync = BluebirdPromise.promisify<any, any>(this.collection.insert, { context: this.collection });
    return insertAsync(document);
  }

  count(query: any): BluebirdPromise<number> {
    const countAsync = BluebirdPromise.promisify<any, any>(this.collection.count, { context: this.collection });
    return countAsync(query);
  }
}