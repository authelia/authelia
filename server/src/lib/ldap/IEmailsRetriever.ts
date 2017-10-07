import BluebirdPromise = require("bluebird");
import { IClient } from "./IClient";

export interface IEmailsRetriever {
  retrieve(username: string, client?: IClient): BluebirdPromise<string[]>;
}