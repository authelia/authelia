import Winston = require("winston");

import { IConnectorFactory } from "./connector/IConnectorFactory";
import { ISessionFactory } from "./ISessionFactory";
import { ISession } from "./ISession";
import { LdapConfiguration } from "../../../configuration/schema/LdapConfiguration";
import { Session } from "./Session";
import { SafeSession } from "./SafeSession";


export class SessionFactory implements ISessionFactory {
  private config: LdapConfiguration;
  private connectorFactory: IConnectorFactory;
  private logger: typeof Winston;

  constructor(ldapConfiguration: LdapConfiguration,
    connectorFactory: IConnectorFactory,
    logger: typeof Winston) {
    this.config = ldapConfiguration;
    this.connectorFactory = connectorFactory;
    this.logger = logger;
  }

  create(userDN: string, password: string): ISession {
    const connector = this.connectorFactory.create();
    return new SafeSession(
      new Session(
        userDN,
        password,
        this.config,
        connector,
        this.logger
      ),
      this.logger
    );
  }
}
