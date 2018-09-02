import BluebirdPromise = require("bluebird");
import exceptions = require("../../../Exceptions");
import { EventEmitter } from "events";
import { ISession } from "./ISession";
import { LdapConfiguration } from "../../../configuration/schema/LdapConfiguration";
import { Winston } from "../../../../../types/Dependencies";
import Util = require("util");
import { HashGenerator } from "../../../utils/HashGenerator";
import { IConnector } from "./connector/IConnector";
import { UsersWithNetworkAddresses } from "../UsersWithNetworkAddresses";

export class Session implements ISession {
  private userDN: string;
  private password: string;
  private connector: IConnector;
  private logger: Winston;
  private options: LdapConfiguration;

  private groupsSearchBase: string;
  private usersSearchBase: string;

  constructor(userDN: string, password: string, options: LdapConfiguration,
    connector: IConnector, logger: Winston) {
    this.options = options;
    this.logger = logger;
    this.userDN = userDN;
    this.password = password;
    this.connector = connector;

    this.groupsSearchBase = (this.options.additional_groups_dn)
      ? Util.format("%s,%s", this.options.additional_groups_dn, this.options.base_dn)
      : this.options.base_dn;

    this.usersSearchBase = (this.options.additional_users_dn)
      ? Util.format("%s,%s", this.options.additional_users_dn, this.options.base_dn)
      : this.options.base_dn;
  }

  open(): BluebirdPromise<void> {
    this.logger.debug("LDAP: Bind user '%s'", this.userDN);
    return this.connector.bindAsync(this.userDN, this.password)
      .error(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapBindError(err.message));
      });
  }

  close(): BluebirdPromise<void> {
    this.logger.debug("LDAP: Unbind user '%s'", this.userDN);
    return this.connector.unbindAsync()
      .error(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapBindError(err.message));
      });
  }

  private createGroupsFilter(userGroupsFilter: string, username: string): BluebirdPromise<string> {
    if (userGroupsFilter.indexOf("{0}") > 0) {
      return BluebirdPromise.resolve(userGroupsFilter.replace("{0}", username));
    }
    else if (userGroupsFilter.indexOf("{dn}") > 0) {
      return this.searchUserDn(username)
        .then(function (userDN: string) {
          return BluebirdPromise.resolve(userGroupsFilter.replace("{dn}", userDN));
        });
    }
    return BluebirdPromise.resolve(userGroupsFilter);
  }

  searchGroups(username: string): BluebirdPromise<string[]> {
    const that = this;
    return this.createGroupsFilter(this.options.groups_filter, username)
      .then(function (groupsFilter: string) {
        that.logger.debug("Computed groups filter is %s", groupsFilter);
        const query = {
          scope: "sub",
          attributes: [that.options.group_name_attribute],
          filter: groupsFilter
        };
        return that.connector.searchAsync(that.groupsSearchBase, query);
      })
      .then(function (docs: { cn: string }[]) {
        const groups = docs.map((doc: any) => { return doc.cn; });
        that.logger.debug("LDAP: groups of user %s are [%s]", username, groups.join(","));
        return BluebirdPromise.resolve(groups);
      });
  }

  searchUserDn(username: string): BluebirdPromise<string> {
    const that = this;
    const filter = this.options.users_filter.replace("{0}", username);
    this.logger.debug("Computed users filter is %s", filter);
    const query = {
      scope: "sub",
      sizeLimit: 1,
      attributes: ["dn"],
      filter: filter
    };

    that.logger.debug("LDAP: searching for user dn of %s", username);
    return that.connector.searchAsync(this.usersSearchBase, query)
      .then(function (users: { dn: string }[]) {
        if (users.length > 0) {
          that.logger.debug("LDAP: retrieved user dn is %s", users[0].dn);
          return BluebirdPromise.resolve(users[0].dn);
        }
        return BluebirdPromise.reject(new Error(
          Util.format("No user DN found for user '%s'", username)));
      });
  }

  searchWhitelist(): BluebirdPromise<UsersWithNetworkAddresses[]> {
    const that = this;
    const users_filter = this.options.users_filter.substr(0, this.options.users_filter.indexOf("="));

    const query = {
      scope: "sub",
      attributes: [users_filter, this.options.whitelist_attribute],
      filter: `(${this.options.whitelist_attribute}=*)`,
    };

    return that.connector.searchAsync(that.usersSearchBase, query)
      .then((users) => {
        const normalisedUsers = users.map((user) => {
          return {
            user: user[users_filter],
            network_addresses: user[that.options.whitelist_attribute],
          };
        });
        return BluebirdPromise.resolve(normalisedUsers);
      })
      .catch((err: Error) => {
        return BluebirdPromise.reject(new exceptions.LdapError("Error while searching whitelist. " + err.stack));
      });
  }

  searchEmails(username: string): BluebirdPromise<string[]> {
    const that = this;
    const query = {
      scope: "base",
      sizeLimit: 1,
      attributes: [this.options.mail_attribute]
    };

    return this.searchUserDn(username)
      .then(function (userDN) {
        return that.connector.searchAsync(userDN, query);
      })
      .then(function (docs: { [mail_attribute: string]: string }[]) {
        const emails: string[] = docs
          .filter((d) => { return typeof d[that.options.mail_attribute] === "string"; })
          .map((d) => { return d[that.options.mail_attribute]; });
        that.logger.debug("LDAP: emails of user '%s' are %s", username, emails);
        return BluebirdPromise.resolve(emails);
      })
      .catch(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapError("Error while searching emails. " + err.stack));
      });
  }

  modifyPassword(username: string, newPassword: string): BluebirdPromise<void> {
    const that = this;
    this.logger.debug("LDAP: update password of user '%s'", username);
    return this.searchUserDn(username)
      .then(function (userDN: string) {
        return BluebirdPromise.join(
          HashGenerator.ssha512(newPassword),
          BluebirdPromise.resolve(userDN));
      })
      .then(function (res: string[]) {
        const change = {
          operation: "replace",
          modification: {
            userPassword: res[0]
          }
        };
        that.logger.debug("Password new='%s'", change.modification.userPassword);
        return that.connector.modifyAsync(res[1], change);
      })
      .then(function () {
        return that.connector.unbindAsync();
      });
  }
}
