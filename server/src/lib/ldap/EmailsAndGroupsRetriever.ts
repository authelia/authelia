import BluebirdPromise = require("bluebird");
import exceptions = require("../Exceptions");
import ldapjs = require("ldapjs");
import { Client } from "./Client";
import { IClientFactory } from "./IClientFactory";
import { LdapConfiguration } from "../configuration/schema/LdapConfiguration";
import { GroupsAndEmails } from "./IClient";


export class EmailsAndGroupsRetriever {
  private options: LdapConfiguration;
  private clientFactory: IClientFactory;

  constructor(options: LdapConfiguration, clientFactory: IClientFactory) {
    this.options = options;
    this.clientFactory = clientFactory;
  }

  retrieve(username: string): BluebirdPromise<GroupsAndEmails> {
    const adminClient = this.clientFactory.create(this.options.user, this.options.password);
    let emails: string[];
    let groups: string[];

    return adminClient.open()
      .then(function () {
        return adminClient.searchEmails(username);
      })
      .then(function (emails_: string[]) {
        emails = emails_;
        return adminClient.searchGroups(username);
      })
      .then(function (groups_: string[]) {
        groups = groups_;
        return adminClient.close();
      })
      .then(function () {
        return BluebirdPromise.resolve({
          emails: emails,
          groups: groups
        });
      })
      .error(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapError("Failed during emails and groups retrieval: " + err.message));
      });
  }
}
