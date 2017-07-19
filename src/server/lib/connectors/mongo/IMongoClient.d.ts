import MongoDB = require("mongodb");

export interface IMongoClient {
    collection(name: string): MongoDB.Collection;
}