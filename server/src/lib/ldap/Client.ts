
import util = require("util");
import BluebirdPromise = require("bluebird");
import exceptions = require("../Exceptions");
import Dovehash = require("dovehash");

import { EventEmitter } from "events";
import { IClient, GroupsAndEmails } from "./IClient";
import { ILdapClient } from "./ILdapClient";
import { ILdapClientFactory } from "./ILdapClientFactory";
import { LdapConfiguration } from "../configuration/Configuration";
import { Winston } from "../../../types/Dependencies";
import Util = require("util");


export class Client implements IClient {
  private userDN: string;
  private password: string;
  private ldapClient: ILdapClient;
  private logger: Winston;
  private dovehash: typeof Dovehash;
  private options: LdapConfiguration;

  constructor(userDN: string, password: string, options: LdapConfiguration,
    ldapClientFactory: ILdapClientFactory, dovehash: typeof Dovehash, logger: Winston) {
    this.options = options;
    this.dovehash = dovehash;
    this.logger = logger;
    this.userDN = userDN;
    this.password = password;
    this.ldapClient = ldapClientFactory.create();
  }

  open(): BluebirdPromise<void> {
    this.logger.debug("LDAP: Bind user '%s'", this.userDN);
    return this.ldapClient.bindAsync(this.userDN, this.password)
      .error(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapBindError(err.message));
      });
  }

  close(): BluebirdPromise<void> {
    this.logger.debug("LDAP: Unbind user '%s'", this.userDN);
    return this.ldapClient.unbindAsync()
      .error(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapBindError(err.message));
      });
  }

  searchGroups(username: string): BluebirdPromise<string[]> {
    const that = this;
    const filter = that.options.groups_filter.replace("{0}", username);
    const query = {
      scope: "sub",
      attributes: [that.options.group_name_attribute],
      filter: filter
    };
    return this.ldapClient.searchAsync(that.options.groups_dn, query)
      .then(function (docs: { cn: string }[]) {
        const groups = docs.map((doc: any) => { return doc.cn; });
        that.logger.debug("LDAP: groups of user %s are %s", username, groups);
        return BluebirdPromise.resolve(groups);
      });
  }

  searchUserDn(username: string): BluebirdPromise<string> {
    const that = this;
    const filter = this.options.users_filter.replace("{0}", username);
    const query = {
      scope: "sub",
      sizeLimit: 1,
      attributes: ["dn"],
      filter: filter
    };

    that.logger.debug("LDAP: searching for user dn of %s", username);
    return that.ldapClient.searchAsync(this.options.users_dn, query)
      .then(function (users: { dn: string }[]) {
        if (users.length > 0) {
          that.logger.debug("LDAP: retrieved user dn is %s", users[0].dn);
          return BluebirdPromise.resolve(users[0].dn);
        }
        return BluebirdPromise.reject(new Error(
          Util.format("No user DN found for user '%s'", username)));
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
        return that.ldapClient.searchAsync(userDN, query);
      })
      .then(function (docs: { mail: string }[]) {
        const emails: string[] = docs
          .filter((d) => { return typeof d.mail === "string"; })
          .map((d) => { return d.mail; });
        that.logger.debug("LDAP: emails of user '%s' are %s", username, emails);
        return BluebirdPromise.resolve(emails);
      })
      .catch(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapError("Error while searching emails. " + err.stack));
      });
  }

  modifyPassword(username: string, newPassword: string): BluebirdPromise<void> {
    const that = this;
    const encodedPassword = this.dovehash.encode("SSHA", newPassword);
    const change = {
      operation: "replace",
      modification: {
        userPassword: encodedPassword
      }
    };

    this.logger.debug("LDAP: update password of user '%s'", username);
    return this.searchUserDn(username)
      .then(function (userDN: string) {
        that.ldapClient.modifyAsync(userDN, change);
      })
      .then(function () {
        return that.ldapClient.unbindAsync();
      });
  }
}
