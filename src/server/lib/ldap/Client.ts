
import util = require("util");
import BluebirdPromise = require("bluebird");
import exceptions = require("../Exceptions");
import ldapjs = require("ldapjs");
import { buildUserDN } from "./common";

import { EventEmitter } from "events";
import { LdapConfiguration } from "../configuration/Configuration";
import { Winston, Ldapjs, Dovehash } from "../../../types/Dependencies";

interface SearchEntry {
  object: any;
}

export interface Attributes {
  groups: string[];
  emails: string[];
}

export class Client {
  private userDN: string;
  private password: string;
  private client: ldapjs.ClientAsync;

  private ldapjs: Ldapjs;
  private logger: Winston;
  private dovehash: Dovehash;
  private options: LdapConfiguration;

  constructor(userDN: string, password: string, options: LdapConfiguration, ldapjs: Ldapjs, dovehash: Dovehash, logger: Winston) {
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

    this.client = BluebirdPromise.promisifyAll(ldapClient) as ldapjs.ClientAsync;
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

  private search(base: string, query: ldapjs.SearchOptions): BluebirdPromise<any> {
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
    const userDN = buildUserDN(username, this.options);
    const password = this.options.password;

    let groupNameAttribute = this.options.group_name_attribute;
    if (!groupNameAttribute) groupNameAttribute = "cn";

    const additionalGroupDN = this.options.additional_group_dn;
    const base_dn = this.options.base_dn;

    let groupDN = base_dn;
    if (additionalGroupDN)
      groupDN = util.format("%s,", additionalGroupDN) + groupDN;

    const query = {
      scope: "sub",
      attributes: [groupNameAttribute],
      filter: "member=" + userDN
    };

    const groups: string[] = [];
    return that.search(groupDN, query)
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

  searchEmails(username: string): BluebirdPromise<string[]> {
    const that = this;
    const userDN = buildUserDN(username, this.options);

    const query = {
      scope: "base",
      sizeLimit: 1,
      attributes: ["mail"]
    };

    return this.search(userDN, query)
      .then(function (docs) {
        const emails = [];
        for (let i = 0; i < docs.length; ++i) {
          if (typeof docs[i].mail === "string")
            emails.push(docs[i].mail);
          else {
            emails.concat(docs[i].mail);
          }
        }
        that.logger.debug("LDAP: emails of user '%s' are %s", username, emails);
        return BluebirdPromise.resolve(emails);
      });
  }

  searchEmailsAndGroups(username: string): BluebirdPromise<Attributes> {
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
    const userDN = buildUserDN(username, this.options);

    const encodedPassword = this.dovehash.encode("SSHA", newPassword);
    const change = {
      operation: "replace",
      modification: {
        userPassword: encodedPassword
      }
    };

    this.logger.debug("LDAP: update password of user '%s'", username);
    return this.client.modifyAsync(userDN, change)
      .then(function () {
        return that.client.unbindAsync();
      });
  }
}
