import BluebirdPromise = require("bluebird");
import { IMongoClient } from "./IMongoClient";

export interface IMongoConnector {
    connect(): BluebirdPromise<IMongoClient>;
}