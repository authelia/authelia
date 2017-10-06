
import util = require("util");
import BluebirdPromise = require("bluebird");
import exceptions = require("../Exceptions");
import Ldapjs = require("ldapjs");
import Dovehash = require("dovehash");

import { EventEmitter } from "events";
import { IClient, GroupsAndEmails } from "./IClient";
import { LdapConfiguration } from "../configuration/Configuration";
import { Winston } from "../../../types/Dependencies";

interface SearchEntry {
  object: any;
}

declare module "ldapjs" {
    export interface ClientAsync {
        on(event: string, callback: (data?: any) => void): void;
        bindAsync(username: string, password: string): BluebirdPromise<void>;
        unbindAsync(): BluebirdPromise<void>;
        searchAsync(base: string, query: Ldapjs.SearchOptions): BluebirdPromise<EventEmitter>;
        modifyAsync(userdn: string, change: Ldapjs.Change): BluebirdPromise<void>;
    }
}

export class Client implements IClient {
  private userDN: string;
  private password: string;
  private client: Ldapjs.ClientAsync;

  private ldapjs: typeof Ldapjs;
  private logger: Winston;
  private dovehash: typeof Dovehash;
  private options: LdapConfiguration;

  constructor(userDN: string, password: string, options: LdapConfiguration,
    ldapjs: typeof Ldapjs, dovehash: typeof Dovehash, logger: Winston) {
    this.options = options;
    this.ldapjs = ldapjs;
    this.dovehash = dovehash;
    this.logger = logger;
    this.userDN = userDN;
    this.password = password;

    const ldapClient = ldapjs.createClient({
      url: this.options.url,
      reconnect: true
    });

    /*const clientLogger = (ldapClient as any).log;
    if (clientLogger) {
      clientLogger.level("trace");
    }*/

    this.client = BluebirdPromise.promisifyAll(ldapClient) as Ldapjs.ClientAsync;
  }

  open(): BluebirdPromise<void> {
    this.logger.debug("LDAP: Bind user '%s'", this.userDN);
    return this.client.bindAsync(this.userDN, this.password)
      .error(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapBindError(err.message));
      });
  }

  close(): BluebirdPromise<void> {
    this.logger.debug("LDAP: Unbind user '%s'", this.userDN);
    return this.client.unbindAsync()
      .error(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapBindError(err.message));
      });
  }

  private search(base: string, query: Ldapjs.SearchOptions): BluebirdPromise<any> {
    const that = this;

    that.logger.debug("LDAP: Search for '%s' in '%s'", JSON.stringify(query), base);
    return that.client.searchAsync(base, query)
      .then(function (res: EventEmitter) {
        const doc: SearchEntry[] = [];

        return new BluebirdPromise((resolve, reject) => {
          res.on("searchEntry", function (entry: SearchEntry) {
            that.logger.debug("Entry retrieved from LDAP is '%s'", JSON.stringify(entry.object));
            doc.push(entry.object);
          });
          res.on("error", function (err: Error) {
            that.logger.error("LDAP: Error received during search '%s'.", JSON.stringify(err));
            reject(new exceptions.LdapSearchError(err.message));
          });
          res.on("end", function () {
            that.logger.debug("LDAP: Search ended and results are '%s'.", JSON.stringify(doc));
            resolve(doc);
          });
        });
      })
      .catch(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapSearchError(err.message));
      });
  }

  private searchGroups(username: string): BluebirdPromise<string[]> {
    const that = this;

    const groups: string[] = [];
    return that.searchUserDn(username)
      .then(function (userDN: string) {
        const filter = that.options.groups_filter.replace("{0}", userDN);
        const query = {
          scope: "sub",
          attributes: [that.options.group_name_attribute],
          filter: filter
        };
        return that.search(that.options.groups_dn, query);
      })
      .then(function (docs) {
        for (let i = 0; i < docs.length; ++i) {
          groups.push(docs[i].cn);
        }
        that.logger.debug("LDAP: groups of user %s are %s", username, groups);
      })
      .then(function () {
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
    return that.search(this.options.users_dn, query)
      .then(function (users: { dn: string }[]) {
        that.logger.debug("LDAP: retrieved user dn is %s", users[0].dn);
        return BluebirdPromise.resolve(users[0].dn);
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
        return that.search(userDN, query);
      })
      .then(function (docs: { mail: string }[]) {
        const emails: string[] = [];
        if (typeof docs[0].mail === "string")
          emails.push(docs[0].mail);
        else {
          emails.concat(docs[0].mail);
        }
        that.logger.debug("LDAP: emails of user '%s' are %s", username, emails);
        return BluebirdPromise.resolve(emails);
      });
  }

  searchEmailsAndGroups(username: string): BluebirdPromise<GroupsAndEmails> {
    const that = this;
    let retrievedEmails: string[], retrievedGroups: string[];

    return this.searchEmails(username)
      .then(function (emails: string[]) {
        retrievedEmails = emails;
        return that.searchGroups(username);
      })
      .then(function (groups: string[]) {
        retrievedGroups = groups;
        return BluebirdPromise.resolve({
          emails: retrievedEmails,
          groups: retrievedGroups
        });
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
        that.client.modifyAsync(userDN, change);
      })
      .then(function () {
        return that.client.unbindAsync();
      });
  }
}
