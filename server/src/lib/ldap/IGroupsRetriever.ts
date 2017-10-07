import BluebirdPromise = require("bluebird");
import { IClient } from "./IClient";

export interface IGroupsRetriever {
  retrieve(username: string): BluebirdPromise<string[]>;
}