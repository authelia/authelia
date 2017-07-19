import BluebirdPromise = require("bluebird");
import exceptions = require("../Exceptions");
import ldapjs = require("ldapjs");
import { Client } from "./Client";
import { buildUserDN } from "./common";

import { LdapConfiguration } from "../configuration/Configuration";
import { Winston, Ldapjs, Dovehash } from "../../../types/Dependencies";


export class EmailsRetriever {
  private options: LdapConfiguration;
  private ldapjs: Ldapjs;
  private logger: Winston;

  constructor(options: LdapConfiguration, ldapjs: Ldapjs, logger: Winston) {
    this.options = options;
    this.ldapjs = ldapjs;
    this.logger = logger;
  }

  private createClient(userDN: string, password: string): Client {
    return new Client(userDN, password, this.options, this.ldapjs, undefined, this.logger);
  }

  retrieve(username: string): BluebirdPromise<string[]> {
    const userDN = buildUserDN(username, this.options);
    const adminClient = this.createClient(this.options.user, this.options.password);
    let emails: string[];

    return adminClient.open()
      .then(function () {
        return adminClient.searchEmails(username);
      })
      .then(function (emails_: string[]) {
        emails = emails_;
        return adminClient.close();
      })
      .then(function() {
        return BluebirdPromise.resolve(emails);
      })
      .error(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapError("Failed during password update: " + err.message));
      });
  }
}
