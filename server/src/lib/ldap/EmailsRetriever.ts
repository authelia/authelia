import BluebirdPromise = require("bluebird");
import exceptions = require("../Exceptions");
import ldapjs = require("ldapjs");
import { Client } from "./Client";

import { IClientFactory } from "./IClientFactory";
import { IEmailsRetriever } from "./IEmailsRetriever";
import { LdapConfiguration } from "../configuration/schema/LdapConfiguration";


export class EmailsRetriever implements IEmailsRetriever {
  private options: LdapConfiguration;
  private clientFactory: IClientFactory;

  constructor(options: LdapConfiguration, clientFactory: IClientFactory) {
    this.options = options;
    this.clientFactory = clientFactory;
  }

  retrieve(username: string): BluebirdPromise<string[]> {
    const adminClient = this.clientFactory.create(this.options.user, this.options.password);
    let emails: string[];

    return adminClient.open()
      .then(function () {
        return adminClient.searchEmails(username);
      })
      .then(function (emails_: string[]) {
        emails = emails_;
        return adminClient.close();
      })
      .then(function () {
        return BluebirdPromise.resolve(emails);
      })
      .catch(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapError("Failed during email retrieval: " + err.message));
      });
  }
}
