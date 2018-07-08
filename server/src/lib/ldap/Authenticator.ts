import BluebirdPromise = require("bluebird");
import exceptions = require("../Exceptions");
import ldapjs = require("ldapjs");
import { IClient } from "./IClient";
import { IClientFactory } from "./IClientFactory";
import { GroupsAndEmails } from "./IClient";

import { IAuthenticator } from "./IAuthenticator";
import { LdapConfiguration } from "../configuration/schema/LdapConfiguration";
import { EmailsAndGroupsRetriever } from "./EmailsAndGroupsRetriever";


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
    const emailsAndGroupsRetriever = new EmailsAndGroupsRetriever(this.options, this.clientFactory);

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
        return emailsAndGroupsRetriever.retrieve(username);
      })
      .then(function (groupsAndEmails: GroupsAndEmails) {
        return BluebirdPromise.resolve(groupsAndEmails);
      })
      .error(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapError(err.message));
      });
  }
}