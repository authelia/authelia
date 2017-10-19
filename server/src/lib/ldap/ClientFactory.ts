import { IClientFactory } from "./IClientFactory";
import { IClient } from "./IClient";
import { Client } from "./Client";
import { SanitizedClient } from "./SanitizedClient";
import { ILdapClientFactory } from "./ILdapClientFactory";
import { LdapConfiguration } from "../configuration/Configuration";
import Ldapjs = require("ldapjs");
import Winston = require("winston");

export class ClientFactory implements IClientFactory {
  private config: LdapConfiguration;
  private ldapClientFactory: ILdapClientFactory;
  private logger: typeof Winston;

  constructor(ldapConfiguration: LdapConfiguration,
    ldapClientFactory: ILdapClientFactory,
    logger: typeof Winston) {
    this.config = ldapConfiguration;
    this.ldapClientFactory = ldapClientFactory;
    this.logger = logger;
  }

  create(userDN: string, password: string): IClient {
    return new SanitizedClient(new Client(userDN, password,
      this.config, this.ldapClientFactory, this.logger));
  }
}