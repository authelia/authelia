import { ILdapClientFactory } from "./ILdapClientFactory";
import { ILdapClient } from "./ILdapClient";
import { LdapClient } from "./LdapClient";
import { LdapConfiguration } from "../configuration/Configuration";

import Ldapjs = require("ldapjs");

export class LdapClientFactory implements ILdapClientFactory {
  private config: LdapConfiguration;
  private ldapjs: typeof Ldapjs;

  constructor(ldapConfiguration: LdapConfiguration, ldapjs: typeof Ldapjs) {
    this.config = ldapConfiguration;
    this.ldapjs = ldapjs;
  }

  create(): ILdapClient {
    return new LdapClient(this.config.url, this.ldapjs);
  }
}