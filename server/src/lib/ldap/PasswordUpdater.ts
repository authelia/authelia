import BluebirdPromise = require("bluebird");
import exceptions = require("../Exceptions");
import ldapjs = require("ldapjs");
import { Client } from "./Client";

import { IPasswordUpdater } from "./IPasswordUpdater";
import { LdapConfiguration } from "../configuration/Configuration";
import { IClientFactory } from "./IClientFactory";


export class PasswordUpdater implements IPasswordUpdater {
  private options: LdapConfiguration;
  private clientFactory: IClientFactory;

  constructor(options: LdapConfiguration, clientFactory: IClientFactory) {
    this.options = options;
    this.clientFactory = clientFactory;
  }

  updatePassword(username: string, newPassword: string)
    : BluebirdPromise<void> {
    const adminClient = this.clientFactory.create(this.options.user,
      this.options.password);

    return adminClient.open()
      .then(function () {
        return adminClient.modifyPassword(username, newPassword);
      })
      .then(function () {
        return adminClient.close();
      })
      .catch(function (err: Error) {
        return BluebirdPromise.reject(
          new exceptions.LdapError(
            "Error while updating password: " + err.message));
      });
  }
}
