/* istanbul ignore next */
import BluebirdPromise = require("bluebird");

/* istanbul ignore next */
export interface ICollection {
    find(query: any, sortKeys: any, count: number): BluebirdPromise<any>;
    findOne(query: any): BluebirdPromise<any>;
    update(query: any, updateQuery: any, options?: any): BluebirdPromise<any>;
    remove(query: any): BluebirdPromise<any>;
    insert(document: any): BluebirdPromise<any>;
}