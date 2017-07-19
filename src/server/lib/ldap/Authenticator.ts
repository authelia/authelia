import BluebirdPromise = require("bluebird");
import exceptions = require("../Exceptions");
import ldapjs = require("ldapjs");
import { Client, Attributes } from "./Client";
import { buildUserDN } from "./common";

import { LdapConfiguration } from "./../../../types/Configuration";
import { Winston, Ldapjs, Dovehash } from "../../../types/Dependencies";


export class Authenticator {
  private options: LdapConfiguration;
  private ldapjs: Ldapjs;
  private logger: Winston;

  constructor(options: LdapConfiguration, ldapjs: Ldapjs, logger: Winston) {
    this.options = options;
    this.ldapjs = ldapjs;
    this.logger = logger;
  }

  private createClient(userDN: string, password: string): Client {
    return new Client(userDN, password, this.options, this.ldapjs, undefined, this.logger);
  }

  authenticate(username: string, password: string): BluebirdPromise<Attributes> {
    const self = this;
    const userDN = buildUserDN(username, this.options);
    const userClient = this.createClient(userDN, password);
    const adminClient = this.createClient(this.options.user, this.options.password);
    let attributes: Attributes;

    return userClient.open()
      .then(function () {
        return userClient.close();
      })
      .then(function () {
        return adminClient.open();
      })
      .then(function () {
        return adminClient.searchEmailsAndGroups(username);
      })
      .then(function (attr: Attributes) {
        attributes = attr;
        return adminClient.close();
      })
      .then(function () {
        return BluebirdPromise.resolve(attributes);
      })
      .error(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.LdapError(err.message));
      });
  }
}