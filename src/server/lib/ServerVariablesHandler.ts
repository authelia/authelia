
import winston = require("winston");
import BluebirdPromise = require("bluebird");
import { IAuthenticator } from "./ldap/IAuthenticator";
import { IPasswordUpdater } from "./ldap/IPasswordUpdater";
import { IEmailsRetriever } from "./ldap/IEmailsRetriever";
import { Authenticator } from "./ldap/Authenticator";
import { PasswordUpdater } from "./ldap/PasswordUpdater";
import { EmailsRetriever } from "./ldap/EmailsRetriever";
import { ClientFactory } from "./ldap/ClientFactory";

import { TOTPValidator } from "./TOTPValidator";
import { TOTPGenerator } from "./TOTPGenerator";
import U2F = require("u2f");
import { IUserDataStore } from "./storage/IUserDataStore";
import { UserDataStore } from "./storage/UserDataStore";
import { INotifier } from "./notifiers/INotifier";
import { AuthenticationRegulator } from "./AuthenticationRegulator";
import Configuration = require("./configuration/Configuration");
import { AccessController } from "./access_control/AccessController";
import { NotifierFactory } from "./notifiers/NotifierFactory";
import { CollectionFactoryFactory } from "./storage/CollectionFactoryFactory";
import { ICollectionFactory } from "./storage/ICollectionFactory";
import { MongoCollectionFactory } from "./storage/mongo/MongoCollectionFactory";
import { MongoConnectorFactory } from "./connectors/mongo/MongoConnectorFactory";
import { IMongoClient } from "./connectors/mongo/IMongoClient";

import { GlobalDependencies } from "../../types/Dependencies";

import express = require("express");

export const VARIABLES_KEY = "authelia-variables";

export interface ServerVariables {
  logger: typeof winston;
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

class UserDataStoreFactory {
  static create(config: Configuration.AppConfiguration): BluebirdPromise<UserDataStore> {
    if (config.storage.local) {
      const nedbOptions = {
        directory: config.storage.local.path,
        inMemory: config.storage.local.in_memory
      };
      const collectionFactory = CollectionFactoryFactory.createNedb(nedbOptions);
      return BluebirdPromise.resolve(new UserDataStore(collectionFactory));
    }
    else if (config.storage.mongo) {
      const mongoConnectorFactory = new MongoConnectorFactory();
      const mongoConnector = mongoConnectorFactory.create(config.storage.mongo.url);
      return mongoConnector.connect()
        .then(function (client: IMongoClient) {
          const collectionFactory = CollectionFactoryFactory.createMongo(client);
          return BluebirdPromise.resolve(new UserDataStore(collectionFactory));
        });
    }

    return BluebirdPromise.reject(new Error("Storage backend incorrectly configured."));
  }
}

export class ServerVariablesHandler {
  static initialize(app: express.Application, config: Configuration.AppConfiguration, deps: GlobalDependencies): BluebirdPromise<void> {
    const five_minutes = 5 * 60;

    const notifier = NotifierFactory.build(config.notifier, deps.nodemailer);
    const ldapClientFactory = new ClientFactory(config.ldap, deps.ldapjs, deps.dovehash, deps.winston);

    const ldapAuthenticator = new Authenticator(config.ldap, ldapClientFactory);
    const ldapPasswordUpdater = new PasswordUpdater(config.ldap, ldapClientFactory);
    const ldapEmailsRetriever = new EmailsRetriever(config.ldap, ldapClientFactory);
    const accessController = new AccessController(config.access_control, deps.winston);
    const totpValidator = new TOTPValidator(deps.speakeasy);
    const totpGenerator = new TOTPGenerator(deps.speakeasy);

    return UserDataStoreFactory.create(config)
      .then(function (userDataStore: UserDataStore) {
        const regulator = new AuthenticationRegulator(userDataStore, five_minutes);

        const variables: ServerVariables = {
          accessController: accessController,
          config: config,
          ldapAuthenticator: ldapAuthenticator,
          ldapPasswordUpdater: ldapPasswordUpdater,
          ldapEmailsRetriever: ldapEmailsRetriever,
          logger: deps.winston,
          notifier: notifier,
          regulator: regulator,
          totpGenerator: totpGenerator,
          totpValidator: totpValidator,
          u2f: deps.u2f,
          userDataStore: userDataStore
        };

        app.set(VARIABLES_KEY, variables);
      });
  }

  static getLogger(app: express.Application): typeof winston {
    return (app.get(VARIABLES_KEY) as ServerVariables).logger;
  }

  static getUserDataStore(app: express.Application): IUserDataStore {
    return (app.get(VARIABLES_KEY) as ServerVariables).userDataStore;
  }

  static getNotifier(app: express.Application): INotifier {
    return (app.get(VARIABLES_KEY) as ServerVariables).notifier;
  }

  static getLdapAuthenticator(app: express.Application): IAuthenticator {
    return (app.get(VARIABLES_KEY) as ServerVariables).ldapAuthenticator;
  }

  static getLdapPasswordUpdater(app: express.Application): IPasswordUpdater {
    return (app.get(VARIABLES_KEY) as ServerVariables).ldapPasswordUpdater;
  }

  static getLdapEmailsRetriever(app: express.Application): IEmailsRetriever {
    return (app.get(VARIABLES_KEY) as ServerVariables).ldapEmailsRetriever;
  }

  static getConfiguration(app: express.Application): Configuration.AppConfiguration {
    return (app.get(VARIABLES_KEY) as ServerVariables).config;
  }

  static getAuthenticationRegulator(app: express.Application): AuthenticationRegulator {
    return (app.get(VARIABLES_KEY) as ServerVariables).regulator;
  }

  static getAccessController(app: express.Application): AccessController {
    return (app.get(VARIABLES_KEY) as ServerVariables).accessController;
  }

  static getTOTPGenerator(app: express.Application): TOTPGenerator {
    return (app.get(VARIABLES_KEY) as ServerVariables).totpGenerator;
  }

  static getTOTPValidator(app: express.Application): TOTPValidator {
    return (app.get(VARIABLES_KEY) as ServerVariables).totpValidator;
  }

  static getU2F(app: express.Application): typeof U2F {
    return (app.get(VARIABLES_KEY) as ServerVariables).u2f;
  }
}
