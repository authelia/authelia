import BluebirdPromise = require("bluebird");
import exceptions = require("../Exceptions");
import ldapjs = require("ldapjs");
import { IClient } from "./IClient";

import { IClientFactory } from "./IClientFactory";
import { IGroupsRetriever } from "./IGroupsRetriever";
import { LdapConfiguration } from "../configuration/Configuration";


export class GroupsRetriever implements IGroupsRetriever {
  private options: LdapConfiguration;
  private clientFactory: IClientFactory;

  constructor(options: LdapConfiguration, clientFactory: IClientFactory) {
    this.options = options;
    this.clientFactory = clientFactory;
  }

  retrieve(username: string, client?: IClient): BluebirdPromise<string[]> {
    client = this.clientFactory.create(this.options.user, this.options.password);
    let groups: string[];

    return client.open()
      .then(function () {
        return client.searchGroups(username);
      })
      .then(function (groups_: string[]) {
        groups = groups_;
        return client.close();
      })
      .then(function () {
        return BluebirdPromise.resolve(groups);
      })
      .catch(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapError("Failed during groups retrieval: " + err.message));
      });
  }
}
