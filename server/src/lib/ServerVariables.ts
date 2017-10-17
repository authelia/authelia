import { IRequestLogger } from "./logging/IRequestLogger";
import { IAuthenticator } from "./ldap/IAuthenticator";
import { IPasswordUpdater } from "./ldap/IPasswordUpdater";
import { IEmailsRetriever } from "./ldap/IEmailsRetriever";
import { ITotpHandler } from "./authentication/totp/ITotpHandler";
import { IU2fHandler } from "./authentication/u2f/IU2fHandler";
import { IUserDataStore } from "./storage/IUserDataStore";
import { INotifier } from "./notifiers/INotifier";
import { IRegulator } from "./regulation/IRegulator";
import { AppConfiguration } from "./configuration/Configuration";
import { IAccessController } from "./access_control/IAccessController";

export interface ServerVariables {
  logger: IRequestLogger;
  ldapAuthenticator: IAuthenticator;
  ldapPasswordUpdater: IPasswordUpdater;
  ldapEmailsRetriever: IEmailsRetriever;
  totpHandler: ITotpHandler;
  u2f: IU2fHandler;
  userDataStore: IUserDataStore;
  notifier: INotifier;
  regulator: IRegulator;
  config: AppConfiguration;
  accessController: IAccessController;
}