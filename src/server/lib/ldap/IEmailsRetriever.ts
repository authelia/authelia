import BluebirdPromise = require("bluebird");

export interface IEmailsRetriever {
  retrieve(username: string): BluebirdPromise<string[]>;
}