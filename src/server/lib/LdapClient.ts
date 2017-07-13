
import util = require("util");
import BluebirdPromise = require("bluebird");
import exceptions = require("./Exceptions");
import Dovehash = require("dovehash");
import ldapjs = require("ldapjs");

import { EventEmitter } from "events";
import { LdapConfiguration } from "./../../types/Configuration";
import { Ldapjs } from "../../types/Dependencies";
import { Winston } from "../../types/Dependencies";

interface SearchEntry {
  object: any;
}

export class LdapClient {
  private options: LdapConfiguration;
  private ldapjs: Ldapjs;
  private logger: Winston;
  private adminClient: ldapjs.ClientAsync;

  constructor(options: LdapConfiguration, ldapjs: Ldapjs, logger: Winston) {
    this.options = options;
    this.ldapjs = ldapjs;
    this.logger = logger;

    this.connect();
  }

  private createClient(): ldapjs.ClientAsync {
    const ldapClient = this.ldapjs.createClient({
      url: this.options.url,
      reconnect: true
    });

    ldapClient.on("error", function (err: Error) {
      console.error("LDAP Error:", err.message);
    });

    return BluebirdPromise.promisifyAll(ldapClient) as ldapjs.ClientAsync;
  }

  connect(): BluebirdPromise<void> {
    const userDN = this.options.user;
    const password = this.options.password;

    this.adminClient = this.createClient();
    return this.adminClient.bindAsync(userDN, password);
  }

  private buildUserDN(username: string): string {
    let userNameAttribute = this.options.user_name_attribute;
    // if not provided, default to cn
    if (!userNameAttribute) userNameAttribute = "cn";

    const additionalUserDN = this.options.additional_user_dn;
    const base_dn = this.options.base_dn;

    let userDN = util.format("%s=%s", userNameAttribute, username);
    if (additionalUserDN) userDN += util.format(",%s", additionalUserDN);
    userDN += util.format(",%s", base_dn);
    return userDN;
  }

  checkPassword(username: string, password: string): BluebirdPromise<void> {
    const userDN = this.buildUserDN(username);
    const that = this;
    const ldapClient = this.createClient();

    this.logger.debug("LDAP: Check password by binding user '%s'", userDN);
    return ldapClient.bindAsync(userDN, password)
      .then(function () {
            that.logger.debug("LDAP: Unbind user '%s'", userDN);
        return ldapClient.unbindAsync();
      })
      .error(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapBindError(err.message));
      });
  }

  private search(base: string, query: ldapjs.SearchOptions): BluebirdPromise<any> {
    const that = this;

    that.logger.debug("LDAP: Search for '%s' in '%s'", JSON.stringify(query), base);
    return that.adminClient.searchAsync(base, query)
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
            that.logger.debug("LDAP: Result of search is '%s'.", JSON.stringify(doc));
            resolve(doc);
          });
        });
      })
      .catch(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapSearchError(err.message));
      });
  }

  retrieveGroups(username: string): BluebirdPromise<string[]> {
    const userDN = this.buildUserDN(username);
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

    const that = this;
    this.logger.debug("LDAP: get groups of user %s", username);
    const groups: string[] = [];
    return that.search(groupDN, query)
      .then(function (docs) {
        for (let i = 0; i < docs.length; ++i) {
          groups.push(docs[i].cn);
        }
        that.logger.debug("LDAP: got groups '%s'", groups);
      })
      .then(function () {
        return BluebirdPromise.resolve(groups);
      });
  }

  retrieveEmails(username: string): BluebirdPromise<string[]> {
    const that = this;
    const user_dn = this.buildUserDN(username);

    const query = {
      scope: "base",
      sizeLimit: 1,
      attributes: ["mail"]
    };

    this.logger.debug("LDAP: get emails of user '%s'", username);
    return this.search(user_dn, query)
      .then(function (docs) {
        const emails = [];
        for (let i = 0; i < docs.length; ++i) {
          if (typeof docs[i].mail === "string")
            emails.push(docs[i].mail);
          else {
            emails.concat(docs[i].mail);
          }
        }
        that.logger.debug("LDAP: got emails '%s'", emails);
        return BluebirdPromise.resolve(emails);
      });
  }

  updatePassword(username: string, newPassword: string): BluebirdPromise<void> {
    const user_dn = this.buildUserDN(username);

    const encoded_password = Dovehash.encode("SSHA", newPassword);
    const change = {
      operation: "replace",
      modification: {
        userPassword: encoded_password
      }
    };

    const that = this;
    this.logger.debug("LDAP: update password of user '%s'", username);

    that.logger.debug("LDAP: modify password");
    return that.adminClient.modifyAsync(user_dn, change)
      .then(function () {
        return that.adminClient.unbindAsync();
      });
  }
}
