

import U2F = require("u2f");

import { IRequestLogger } from "./logging/IRequestLogger";
import { IAuthenticator } from "./ldap/IAuthenticator";
import { IPasswordUpdater } from "./ldap/IPasswordUpdater";
import { IEmailsRetriever } from "./ldap/IEmailsRetriever";

import { TOTPValidator } from "./TOTPValidator";
import { TOTPGenerator } from "./TOTPGenerator";
import { IUserDataStore } from "./storage/IUserDataStore";
import { INotifier } from "./notifiers/INotifier";
import { AuthenticationRegulator } from "./AuthenticationRegulator";
import Configuration = require("./configuration/Configuration");
import { AccessController } from "./access_control/AccessController";


export interface ServerVariables {
  logger: IRequestLogger;
  ldapAuthenticator: IAuthenticator;
  ldapPasswordUpdater: IPasswordUpdater;
  ldapEmailsRetriever: IEmailsRetriever;
  totpValidator: TOTPValidator;
  totpGenerator: TOTPGenerator;
  u2f: typeof U2F;
  userDataStore: IUserDataStore;
  notifier: INotifier;
  regulator: AuthenticationRegulator;
  config: Configuration.AppConfiguration;
  accessController: AccessController;
}