
import winston = require("winston");
import BluebirdPromise = require("bluebird");
import U2F = require("u2f");
import Nodemailer = require("nodemailer");

import { IRequestLogger } from "./logging/IRequestLogger";
import { RequestLogger } from "./logging/RequestLogger";

import { IAuthenticator } from "./ldap/IAuthenticator";
import { IPasswordUpdater } from "./ldap/IPasswordUpdater";
import { IEmailsRetriever } from "./ldap/IEmailsRetriever";
import { Authenticator } from "./ldap/Authenticator";
import { PasswordUpdater } from "./ldap/PasswordUpdater";
import { EmailsRetriever } from "./ldap/EmailsRetriever";
import { ClientFactory } from "./ldap/ClientFactory";
import { LdapClientFactory } from "./ldap/LdapClientFactory";

import { TotpHandler } from "./authentication/totp/TotpHandler";
import { ITotpHandler } from "./authentication/totp/ITotpHandler";
import { NotifierFactory } from "./notifiers/NotifierFactory";
import { MailSenderBuilder } from "./notifiers/MailSenderBuilder";

import { IUserDataStore } from "./storage/IUserDataStore";
import { UserDataStore } from "./storage/UserDataStore";
import { INotifier } from "./notifiers/INotifier";
import { Regulator } from "./regulation/Regulator";
import { IRegulator } from "./regulation/IRegulator";
import Configuration = require("./configuration/schema/Configuration");
import { AccessController } from "./access_control/AccessController";
import { IAccessController } from "./access_control/IAccessController";
import { CollectionFactoryFactory } from "./storage/CollectionFactoryFactory";
import { ICollectionFactory } from "./storage/ICollectionFactory";
import { MongoCollectionFactory } from "./storage/mongo/MongoCollectionFactory";
import { IMongoClient } from "./connectors/mongo/IMongoClient";

import { GlobalDependencies } from "../../types/Dependencies";
import { ServerVariables } from "./ServerVariables";
import { MethodCalculator } from "./authentication/MethodCalculator";
import { MongoClient } from "./connectors/mongo/MongoClient";
import { IGlobalLogger } from "./logging/IGlobalLogger";

class UserDataStoreFactory {
  static create(config: Configuration.Configuration, globalLogger: IGlobalLogger): BluebirdPromise<UserDataStore> {
    if (config.storage.local) {
      const nedbOptions: Nedb.DataStoreOptions = {
        filename: config.storage.local.path,
        inMemoryOnly: config.storage.local.in_memory
      };
      const collectionFactory = CollectionFactoryFactory.createNedb(nedbOptions);
      return BluebirdPromise.resolve(new UserDataStore(collectionFactory));
    }
    else if (config.storage.mongo) {
      const mongoClient = new MongoClient(
        config.storage.mongo.url,
        config.storage.mongo.database,
        globalLogger);
      const collectionFactory = CollectionFactoryFactory.createMongo(mongoClient);
      return BluebirdPromise.resolve(new UserDataStore(collectionFactory));
    }

    return BluebirdPromise.reject(new Error("Storage backend incorrectly configured."));
  }
}

export class ServerVariablesInitializer {
  static initialize(
    config: Configuration.Configuration,
    globalLogger: IGlobalLogger,
    requestLogger: IRequestLogger,
    deps: GlobalDependencies): BluebirdPromise<ServerVariables> {

    const mailSenderBuilder = new MailSenderBuilder(Nodemailer);
    const notifier = NotifierFactory.build(config.notifier, mailSenderBuilder);
    const ldapClientFactory = new LdapClientFactory(config.ldap, deps.ldapjs);
    const clientFactory = new ClientFactory(config.ldap, ldapClientFactory,
      deps.winston);

    const ldapAuthenticator = new Authenticator(config.ldap, clientFactory);
    const ldapPasswordUpdater = new PasswordUpdater(config.ldap, clientFactory);
    const ldapEmailsRetriever = new EmailsRetriever(config.ldap, clientFactory);
    const accessController = new AccessController(config.access_control, deps.winston);
    const totpHandler = new TotpHandler(deps.speakeasy);

    return UserDataStoreFactory.create(config, globalLogger)
      .then(function (userDataStore: UserDataStore) {
        const regulator = new Regulator(userDataStore, config.regulation.max_retries,
          config.regulation.find_time, config.regulation.ban_time);

        const variables: ServerVariables = {
          accessController: accessController,
          config: config,
          ldapAuthenticator: ldapAuthenticator,
          ldapPasswordUpdater: ldapPasswordUpdater,
          ldapEmailsRetriever: ldapEmailsRetriever,
          logger: requestLogger,
          notifier: notifier,
          regulator: regulator,
          totpHandler: totpHandler,
          u2f: deps.u2f,
          userDataStore: userDataStore
        };
        return BluebirdPromise.resolve(variables);
      });
  }
}
