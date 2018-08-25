import { IConnector } from "./IConnector";
import { Connector } from "./Connector";
import { LdapConfiguration } from "../../../../configuration/schema/LdapConfiguration";
import { Ldapjs } from "Dependencies";

export class ConnectorFactory {
  private configuration: LdapConfiguration;
  private ldapjs: Ldapjs;

  constructor(configuration: LdapConfiguration, ldapjs: Ldapjs) {
    this.configuration = configuration;
    this.ldapjs = ldapjs;
  }

  create(): IConnector {
    return new Connector(this.configuration.url, this.ldapjs);
  }
}