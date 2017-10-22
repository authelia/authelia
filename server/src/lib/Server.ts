import BluebirdPromise = require("bluebird");
import ObjectPath = require("object-path");

import { AccessController } from "./access_control/AccessController";
import { AppConfiguration, UserConfiguration } from "./configuration/Configuration";
import { GlobalDependencies } from "../../types/Dependencies";
import { UserDataStore } from "./storage/UserDataStore";
import { ConfigurationParser } from "./configuration/ConfigurationParser";
import { SessionConfigurationBuilder } from "./configuration/SessionConfigurationBuilder";
import { GlobalLogger } from "./logging/GlobalLogger";
import { RequestLogger } from "./logging/RequestLogger";
import { ServerVariables } from "./ServerVariables";
import { ServerVariablesInitializer } from "./ServerVariablesInitializer";
import { Configurator } from "./web_server/Configurator";

import * as Express from "express";
import * as Path from "path";
import * as http from "http";

function clone(obj: any) {
  return JSON.parse(JSON.stringify(obj));
}

export default class Server {
  private httpServer: http.Server;
  private globalLogger: GlobalLogger;
  private requestLogger: RequestLogger;

  constructor(deps: GlobalDependencies) {
    this.globalLogger = new GlobalLogger(deps.winston);
    this.requestLogger = new RequestLogger(deps.winston);
  }

  private displayConfigurations(userConfiguration: UserConfiguration,
    appConfiguration: AppConfiguration) {
    const displayableUserConfiguration = clone(userConfiguration);
    const displayableAppConfiguration = clone(appConfiguration);
    const STARS = "*****";

    displayableUserConfiguration.ldap.password = STARS;
    displayableUserConfiguration.session.secret = STARS;
    if (displayableUserConfiguration.notifier && displayableUserConfiguration.notifier.email)
      displayableUserConfiguration.notifier.email.password = STARS;
    if (displayableUserConfiguration.notifier && displayableUserConfiguration.notifier.smtp)
      displayableUserConfiguration.notifier.smtp.password = STARS;

    displayableAppConfiguration.ldap.password = STARS;
    displayableAppConfiguration.session.secret = STARS;
    if (displayableAppConfiguration.notifier && displayableAppConfiguration.notifier.email)
      displayableAppConfiguration.notifier.email.password = STARS;
    if (displayableAppConfiguration.notifier && displayableAppConfiguration.notifier.smtp)
      displayableAppConfiguration.notifier.smtp.password = STARS;

    this.globalLogger.debug("User configuration is %s",
      JSON.stringify(displayableUserConfiguration, undefined, 2));
    this.globalLogger.debug("Adapted configuration is %s",
      JSON.stringify(displayableAppConfiguration, undefined, 2));
  }

  private setup(config: AppConfiguration, app: Express.Application, deps: GlobalDependencies): BluebirdPromise<void> {
    const that = this;
    return ServerVariablesInitializer.initialize(config, this.requestLogger, deps)
      .then(function (vars: ServerVariables) {
        Configurator.configure(config, app, vars, deps);
        return BluebirdPromise.resolve();
      });
  }

  private startServer(app: Express.Application, port: number) {
    const that = this;
    return new BluebirdPromise<void>((resolve, reject) => {
      this.httpServer = app.listen(port, function (err: string) {
        that.globalLogger.info("Listening on port %d...", port);
        resolve();
      });
    });
  }

  start(userConfiguration: UserConfiguration, deps: GlobalDependencies)
    : BluebirdPromise<void> {
    const that = this;
    const app = Express();

    const appConfiguration = ConfigurationParser.parse(userConfiguration);

    // by default the level of logs is info
    deps.winston.level = userConfiguration.logs_level;
    this.displayConfigurations(userConfiguration, appConfiguration);

    return this.setup(appConfiguration, app, deps)
      .then(function () {
        return that.startServer(app, appConfiguration.port);
      });
  }

  stop() {
    this.httpServer.close();
  }
}

