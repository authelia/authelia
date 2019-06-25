import Winston = require("winston");
import { IConnector } from "./IConnector";
import { Connector } from "./Connector";
import { LdapConfiguration } from "../../../../configuration/schema/LdapConfiguration";
import { Ldapjs } from "Dependencies";
import { ClientOptions } from "ldapjs";
import * as fs from "fs";

export class ConnectorFactory {
  private configuration: LdapConfiguration;
  private ldapjs: Ldapjs;
  private logger: typeof Winston;

  constructor(configuration: LdapConfiguration, ldapjs: Ldapjs, logger: typeof Winston) {
    this.configuration = configuration;
    this.ldapjs = ldapjs;
    this.logger = logger;
  }

  create(): IConnector {
    const options: ClientOptions = {
      url: this.configuration.url,
      reconnect: this.configuration.reconnect
    };

    if (this.configuration.caCert && (this.configuration.url.toLowerCase().startsWith("ldaps"))) {
      this.logger.info("Reading CA certificate from: %s", this.configuration.caCert);
      options.tlsOptions = {
        ca: [ fs.readFileSync(this.configuration.caCert, "utf-8") ],
      };
    }

    this.logger.debug("Using ldap client options: %s", JSON.stringify(options));
    return new Connector(options, this.ldapjs);
  }
}