import BluebirdPromise = require("bluebird");
import Nodemailer = require("nodemailer");

import { IRequestLogger } from "./logging/IRequestLogger";

import { TotpHandler } from "./authentication/totp/TotpHandler";
import { NotifierFactory } from "./notifiers/NotifierFactory";
import { MailSenderBuilder } from "./notifiers/MailSenderBuilder";
import { LdapUsersDatabase } from "./authentication/backends/ldap/LdapUsersDatabase";
import { ConnectorFactory } from "./authentication/backends/ldap/connector/ConnectorFactory";

import { UserDataStore } from "./storage/UserDataStore";
import { Regulator } from "./regulation/Regulator";
import Configuration = require("./configuration/schema/Configuration");
import { CollectionFactoryFactory } from "./storage/CollectionFactoryFactory";

import { GlobalDependencies } from "../../types/Dependencies";
import { ServerVariables } from "./ServerVariables";
import { MongoClient } from "./connectors/mongo/MongoClient";
import { IGlobalLogger } from "./logging/IGlobalLogger";
import { SessionFactory } from "./authentication/backends/ldap/SessionFactory";
import { IUsersDatabase } from "./authentication/backends/IUsersDatabase";
import { FileUsersDatabase } from "./authentication/backends/file/FileUsersDatabase";
import { Authorizer } from "./authorization/Authorizer";

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
        config.storage.mongo,
        globalLogger);
      const collectionFactory = CollectionFactoryFactory.createMongo(mongoClient);
      return BluebirdPromise.resolve(new UserDataStore(collectionFactory));
    }

    return BluebirdPromise.reject(new Error("Storage backend incorrectly configured."));
  }
}

export class ServerVariablesInitializer {
  static createUsersDatabase(
    config: Configuration.Configuration,
    deps: GlobalDependencies)
    : IUsersDatabase {

    if (config.authentication_backend.ldap) {
      const ldapConfig = config.authentication_backend.ldap;
      return new LdapUsersDatabase(
        new SessionFactory(
          ldapConfig,
          new ConnectorFactory(ldapConfig, deps.ldapjs, deps.winston),
          deps.winston
        ),
        ldapConfig
      );
    }
    else if (config.authentication_backend.file) {
      return new FileUsersDatabase(config.authentication_backend.file);
    }
  }

  static initialize(
    config: Configuration.Configuration,
    globalLogger: IGlobalLogger,
    requestLogger: IRequestLogger,
    deps: GlobalDependencies)
    : BluebirdPromise<ServerVariables> {

    const mailSenderBuilder =
      new MailSenderBuilder(Nodemailer);
    const notifier = NotifierFactory.build(
      config.notifier, mailSenderBuilder);
    const authorizer = new Authorizer(config.access_control, deps.winston);
    const totpHandler = new TotpHandler(deps.speakeasy);
    const usersDatabase = this.createUsersDatabase(
      config, deps);

    return UserDataStoreFactory.create(config, globalLogger)
      .then(function (userDataStore: UserDataStore) {
        const regulator = new Regulator(userDataStore, config.regulation.max_retries,
          config.regulation.find_time, config.regulation.ban_time);

        const variables: ServerVariables = {
          authorizer: authorizer,
          config: config,
          usersDatabase: usersDatabase,
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
