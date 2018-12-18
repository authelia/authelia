import MongoDB = require("mongodb");
import Bluebird = require("bluebird");

export interface IMongoClient {
    collection(name: string): Bluebird<MongoDB.Collection>
}