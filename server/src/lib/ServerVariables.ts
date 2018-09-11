import { IRequestLogger } from "./logging/IRequestLogger";
import { ITotpHandler } from "./authentication/totp/ITotpHandler";
import { IU2fHandler } from "./authentication/u2f/IU2fHandler";
import { IUserDataStore } from "./storage/IUserDataStore";
import { INotifier } from "./notifiers/INotifier";
import { IRegulator } from "./regulation/IRegulator";
import { Configuration } from "./configuration/schema/Configuration";
import { IAccessController } from "./access_control/IAccessController";
import { IUsersDatabase } from "./authentication/backends/IUsersDatabase";
import { IWhitelistHandler } from "./authentication/whitelist/IWhitelistHandler";

export interface ServerVariables {
  logger: IRequestLogger;
  usersDatabase: IUsersDatabase;
  totpHandler: ITotpHandler;
  u2f: IU2fHandler;
  userDataStore: IUserDataStore;
  notifier: INotifier;
  regulator: IRegulator;
  config: Configuration;
  accessController: IAccessController;
  whitelistHandler: IWhitelistHandler;
}