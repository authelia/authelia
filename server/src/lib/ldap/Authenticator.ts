import BluebirdPromise = require("bluebird");
import exceptions = require("../Exceptions");
import ldapjs = require("ldapjs");
import { IClient } from "./IClient";
import { IClientFactory } from "./IClientFactory";
import { GroupsAndEmails } from "./IClient";

import { IAuthenticator } from "./IAuthenticator";
import { LdapConfiguration } from "../configuration/Configuration";
import { Winston, Ldapjs, Dovehash } from "../../../types/Dependencies";


export class Authenticator implements IAuthenticator {
  private options: LdapConfiguration;
  private clientFactory: IClientFactory;

  constructor(options: LdapConfiguration, clientFactory: IClientFactory) {
    this.options = options;
    this.clientFactory = clientFactory;
  }

  authenticate(username: string, password: string): BluebirdPromise<GroupsAndEmails> {
    const that = this;
    let userClient: IClient;
    const adminClient = this.clientFactory.create(this.options.user, this.options.password);
    let groupsAndEmails: GroupsAndEmails;

    return adminClient.open()
      .then(function () {
        return adminClient.searchUserDn(username);
      })
      .then(function (userDN: string) {
        userClient = that.clientFactory.create(userDN, password);
        return userClient.open();
      })
      .then(function () {
        return userClient.close();
      })
      .then(function () {
        return adminClient.open();
      })
      .then(function () {
        return adminClient.searchEmailsAndGroups(username);
      })
      .then(function (gae: GroupsAndEmails) {
        groupsAndEmails = gae;
        return adminClient.close();
      })
      .then(function () {
        return BluebirdPromise.resolve(groupsAndEmails);
      })
      .error(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapError(err.message));
      });
  }
}