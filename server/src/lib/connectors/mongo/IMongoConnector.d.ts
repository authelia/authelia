import BluebirdPromise = require("bluebird");
import { IMongoClient } from "./IMongoClient";

export interface IMongoConnector {
    connect(databaseName: string): BluebirdPromise<IMongoClient>;
    close(): BluebirdPromise<void>;
}