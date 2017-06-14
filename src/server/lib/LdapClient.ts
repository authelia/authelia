
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
  private client: ldapjs.ClientAsync;

  constructor(options: LdapConfiguration, ldapjs: Ldapjs, logger: Winston) {
    this.options = options;
    this.ldapjs = ldapjs;
    this.logger = logger;

    this.connect();
  }

  connect(): void {
    const ldap_client = this.ldapjs.createClient({
      url: this.options.url,
      reconnect: true
    });

    ldap_client.on("error", function (err: Error) {
      console.error("LDAP Error:", err.message);
    });

    this.client = BluebirdPromise.promisifyAll(ldap_client) as ldapjs.ClientAsync;
  }

  private build_user_dn(username: string): string {
    let user_name_attr = this.options.user_name_attribute;
    // if not provided, default to cn
    if (!user_name_attr) user_name_attr = "cn";

    const additional_user_dn = this.options.additional_user_dn;
    const base_dn = this.options.base_dn;

    let user_dn = util.format("%s=%s", user_name_attr, username);
    if (additional_user_dn) user_dn += util.format(",%s", additional_user_dn);
    user_dn += util.format(",%s", base_dn);
    return user_dn;
  }

  bind(username: string, password: string): BluebirdPromise<void> {
    const user_dn = this.build_user_dn(username);

    this.logger.debug("LDAP: Bind user %s", user_dn);
    return this.client.bindAsync(user_dn, password)
      .error(function (err: Error) {
        throw new exceptions.LdapBindError(err.message);
      });
  }

  private search_in_ldap(base: string, query: ldapjs.SearchOptions): BluebirdPromise<any> {
    this.logger.debug("LDAP: Search for %s in %s", JSON.stringify(query), base);
    return new BluebirdPromise((resolve, reject) => {
      this.client.searchAsync(base, query)
        .then(function (res: EventEmitter) {
          const doc: SearchEntry[] = [];
          res.on("searchEntry", function (entry: SearchEntry) {
            doc.push(entry.object);
          });
          res.on("error", function (err: Error) {
            reject(new exceptions.LdapSearchError(err.message));
          });
          res.on("end", function () {
            resolve(doc);
          });
        })
        .catch(function (err: Error) {
          reject(new exceptions.LdapSearchError(err.message));
        });
    });
  }

  get_groups(username: string): BluebirdPromise<string[]> {
    const user_dn = this.build_user_dn(username);

    let group_name_attr = this.options.group_name_attribute;
    if (!group_name_attr) group_name_attr = "cn";

    const additional_group_dn = this.options.additional_group_dn;
    const base_dn = this.options.base_dn;

    let group_dn = base_dn;
    if (additional_group_dn)
      group_dn = util.format("%s,", additional_group_dn) + group_dn;

    const query = {
      scope: "sub",
      attributes: [group_name_attr],
      filter: "member=" + user_dn
    };

    const that = this;
    this.logger.debug("LDAP: get groups of user %s", username);
    return this.search_in_ldap(group_dn, query)
      .then(function (docs) {
        const groups = [];
        for (let i = 0; i < docs.length; ++i) {
          groups.push(docs[i].cn);
        }
        that.logger.debug("LDAP: got groups %s", groups);
        return BluebirdPromise.resolve(groups);
      });
  }

  get_emails(username: string): BluebirdPromise<string[]> {
    const that = this;
    const user_dn = this.build_user_dn(username);

    const query = {
      scope: "base",
      sizeLimit: 1,
      attributes: ["mail"]
    };

    this.logger.debug("LDAP: get emails of user %s", username);
    return this.search_in_ldap(user_dn, query)
      .then(function (docs) {
        const emails = [];
        for (let i = 0; i < docs.length; ++i) {
          if (typeof docs[i].mail === "string")
            emails.push(docs[i].mail);
          else {
            emails.concat(docs[i].mail);
          }
        }
        that.logger.debug("LDAP: got emails %s", emails);
        return BluebirdPromise.resolve(emails);
      });
  }

  update_password(username: string, new_password: string): BluebirdPromise<void> {
    const user_dn = this.build_user_dn(username);

    const encoded_password = Dovehash.encode("SSHA", new_password);
    const change = {
      operation: "replace",
      modification: {
        userPassword: encoded_password
      }
    };

    const that = this;
    this.logger.debug("LDAP: update password of user %s", username);

    this.logger.debug("LDAP: bind admin");
    return this.client.bindAsync(this.options.user, this.options.password)
      .then(function () {
        that.logger.debug("LDAP: modify password");
        return that.client.modifyAsync(user_dn, change);
      });
  }
}
