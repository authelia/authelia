import BluebirdPromise = require("bluebird");
import exceptions = require("../Exceptions");
import ldapjs = require("ldapjs");
import { Client } from "./Client";
import { buildUserDN } from "./common";

import { LdapConfiguration } from "../configuration/Configuration";
import { Winston, Ldapjs, Dovehash } from "../../../types/Dependencies";


export class PasswordUpdater {
  private options: LdapConfiguration;
  private ldapjs: Ldapjs;
  private logger: Winston;
  private dovehash: Dovehash;

  constructor(options: LdapConfiguration, ldapjs: Ldapjs, dovehash: Dovehash, logger: Winston) {
    this.options = options;
    this.ldapjs = ldapjs;
    this.logger = logger;
    this.dovehash = dovehash;
  }

  private createClient(userDN: string, password: string): Client {
    return new Client(userDN, password, this.options, this.ldapjs, this.dovehash, this.logger);
  }

  updatePassword(username: string, newPassword: string): BluebirdPromise<void> {
    const userDN = buildUserDN(username, this.options);
    const adminClient = this.createClient(this.options.user, this.options.password);

    return adminClient.open()
      .then(function () {
        return adminClient.modifyPassword(username, newPassword);
      })
      .then(function () {
        return adminClient.close();
      })
      .error(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapError("Failed during password update: " + err.message));
      });
  }
}
