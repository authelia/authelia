import { IClientFactory } from "./IClientFactory";
import { IClient } from "./IClient";
import { Client } from "./Client";
import { ILdapClientFactory } from "./ILdapClientFactory";
import { LdapConfiguration } from "../configuration/Configuration";

import Ldapjs = require("ldapjs");
import Dovehash = require("dovehash");
import Winston = require("winston");

export class ClientFactory implements IClientFactory {
  private config: LdapConfiguration;
  private ldapClientFactory: ILdapClientFactory;
  private dovehash: typeof Dovehash;
  private logger: typeof Winston;

  constructor(ldapConfiguration: LdapConfiguration, ldapClientFactory: ILdapClientFactory,
    dovehash: typeof Dovehash, logger: typeof Winston) {
    this.config = ldapConfiguration;
    this.ldapClientFactory = ldapClientFactory;
    this.dovehash = dovehash;
    this.logger = logger;
  }

  create(userDN: string, password: string): IClient {
    return new Client(userDN, password, this.config, this.ldapClientFactory,
      this.dovehash, this.logger);
  }
}