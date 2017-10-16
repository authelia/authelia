import { ServerVariables } from "../../src/lib/ServerVariables";

import { AppConfiguration } from "../../src/lib/configuration/Configuration";
import { AuthenticatorStub } from "./ldap/AuthenticatorStub";
import { EmailsRetrieverStub } from "./ldap/EmailsRetrieverStub";
import { PasswordUpdaterStub } from "./ldap/PasswordUpdaterStub";
import { AccessControllerStub } from "./AccessControllerStub";
import { RequestLoggerStub } from "./RequestLoggerStub";
import { NotifierStub } from "./NotifierStub";
import { RegulatorStub } from "./RegulatorStub";
import { TotpHandlerStub } from "./TotpHandlerStub";
import { UserDataStoreStub } from "./storage/UserDataStoreStub";
import { U2fHandlerStub } from "./U2fHandlerStub";

export interface ServerVariablesMock {
  accessController: AccessControllerStub;
  config: AppConfiguration;
  ldapAuthenticator: AuthenticatorStub;
  ldapEmailsRetriever: EmailsRetrieverStub;
  ldapPasswordUpdater: PasswordUpdaterStub;
  logger: RequestLoggerStub;
  notifier: NotifierStub;
  regulator: RegulatorStub;
  totpHandler: TotpHandlerStub;
  userDataStore: UserDataStoreStub;
  u2f: U2fHandlerStub;
}

export class ServerVariablesMockBuilder {
  static build(): { variables: ServerVariables, mocks: ServerVariablesMock} {
    const mocks: ServerVariablesMock = {
      accessController: new AccessControllerStub(),
      config: {
        access_control: {},
        authentication_methods: {
          default_method: "two_factor"
        },
        ldap: {
          url: "ldap://ldap",
          user: "user",
          password: "password",
          mail_attribute: "mail",
          users_dn: "ou=users,dc=example,dc=com",
          groups_dn: "ou=groups,dc=example,dc=com",
          users_filter: "cn={0}",
          groups_filter: "member={dn}",
          group_name_attribute: "cn"
        },
        logs_level: "debug",
        notifier: {},
        port: 8080,
        regulation: {
          ban_time: 50,
          find_time: 50,
          max_retries: 3
        },
        session: {
          secret: "my_secret"
        },
        storage: {}
      },
      ldapAuthenticator: new AuthenticatorStub(),
      ldapEmailsRetriever: new EmailsRetrieverStub(),
      ldapPasswordUpdater: new PasswordUpdaterStub(),
      logger: new RequestLoggerStub(),
      notifier: new NotifierStub(),
      regulator: new RegulatorStub(),
      totpHandler: new TotpHandlerStub(),
      userDataStore: new UserDataStoreStub(),
      u2f: new U2fHandlerStub()
    };
    const vars: ServerVariables = {
      accessController: mocks.accessController,
      config: mocks.config,
      ldapAuthenticator: mocks.ldapAuthenticator,
      ldapEmailsRetriever: mocks.ldapEmailsRetriever,
      ldapPasswordUpdater: mocks.ldapPasswordUpdater,
      logger: mocks.logger,
      notifier: mocks.notifier,
      regulator: mocks.regulator,
      totpHandler: mocks.totpHandler,
      userDataStore: mocks.userDataStore,
      u2f: mocks.u2f
    };

    return {
      variables: vars,
      mocks: mocks
    };
  }
}