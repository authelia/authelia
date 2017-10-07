
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

import { TOTPValidator } from "./TOTPValidator";
import { TOTPGenerator } from "./TOTPGenerator";

import { NotifierFactory } from "./notifiers/NotifierFactory";
import { MailSenderBuilder } from "./notifiers/MailSenderBuilder";

import { IUserDataStore } from "./storage/IUserDataStore";
import { UserDataStore } from "./storage/UserDataStore";
import { INotifier } from "./notifiers/INotifier";
import { AuthenticationRegulator } from "./AuthenticationRegulator";
import Configuration = require("./configuration/Configuration");
import { AccessController } from "./access_control/AccessController";
import { CollectionFactoryFactory } from "./storage/CollectionFactoryFactory";
import { ICollectionFactory } from "./storage/ICollectionFactory";
import { MongoCollectionFactory } from "./storage/mongo/MongoCollectionFactory";
import { MongoConnectorFactory } from "./connectors/mongo/MongoConnectorFactory";
import { IMongoClient } from "./connectors/mongo/IMongoClient";
import { GlobalDependencies } from "../../types/Dependencies";
import { ServerVariables } from "./ServerVariables";

import express = require("express");

export const VARIABLES_KEY = "authelia-variables";

class UserDataStoreFactory {
  static create(config: Configuration.AppConfiguration): BluebirdPromise<UserDataStore> {
    if (config.storage.local) {
      const nedbOptions: Nedb.DataStoreOptions = {
        filename: config.storage.local.path,
        inMemoryOnly: config.storage.local.in_memory
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
  static initialize(app: express.Application, config: Configuration.AppConfiguration, requestLogger: IRequestLogger,
    deps: GlobalDependencies): BluebirdPromise<void> {
    const mailSenderBuilder = new MailSenderBuilder(Nodemailer);
    const notifier = NotifierFactory.build(config.notifier, mailSenderBuilder);
    const ldapClientFactory = new LdapClientFactory(config.ldap, deps.ldapjs);
    const clientFactory = new ClientFactory(config.ldap, ldapClientFactory, deps.dovehash, deps.winston);

    const ldapAuthenticator = new Authenticator(config.ldap, clientFactory);
    const ldapPasswordUpdater = new PasswordUpdater(config.ldap, clientFactory);
    const ldapEmailsRetriever = new EmailsRetriever(config.ldap, clientFactory);
    const accessController = new AccessController(config.access_control, deps.winston);
    const totpValidator = new TOTPValidator(deps.speakeasy);
    const totpGenerator = new TOTPGenerator(deps.speakeasy);

    return UserDataStoreFactory.create(config)
      .then(function (userDataStore: UserDataStore) {
        const regulator = new AuthenticationRegulator(userDataStore, config.regulation.max_retries,
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
          totpGenerator: totpGenerator,
          totpValidator: totpValidator,
          u2f: deps.u2f,
          userDataStore: userDataStore,
        };

        app.set(VARIABLES_KEY, variables);
      });
  }

  static getLogger(app: express.Application): IRequestLogger {
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
