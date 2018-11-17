import BluebirdPromise = require("bluebird");
import ObjectPath = require("object-path");

import { Configuration } from "./configuration/schema/Configuration";
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

  private displayConfigurations(configuration: Configuration) {
    const displayableConfiguration: Configuration = clone(configuration);
    const STARS = "*****";

    if (displayableConfiguration.authentication_backend.ldap) {
      displayableConfiguration.authentication_backend.ldap.password = STARS;
    }

    displayableConfiguration.session.secret = STARS;
    if (displayableConfiguration.notifier && displayableConfiguration.notifier.email)
      displayableConfiguration.notifier.email.password = STARS;
    if (displayableConfiguration.notifier && displayableConfiguration.notifier.smtp)
      displayableConfiguration.notifier.smtp.password = STARS;

    this.globalLogger.debug("User configuration is %s",
      JSON.stringify(displayableConfiguration, undefined, 2));
  }

  private setup(config: Configuration, app: Express.Application, deps: GlobalDependencies): BluebirdPromise<void> {
    const that = this;
    return ServerVariablesInitializer.initialize(
      config, this.globalLogger, this.requestLogger, deps)
      .then(function (vars: ServerVariables) {
        Configurator.configure(config, app, vars, deps);
        return BluebirdPromise.resolve();
      });
  }

  private startServer(app: Express.Application, port: number) {
    const that = this;
    that.globalLogger.info("Starting Authelia...");
    return new BluebirdPromise<void>((resolve, reject) => {
      this.httpServer = app.listen(port, function (err: string) {
        that.globalLogger.info("Listening on port %d...", port);
        resolve();
      });
    });
  }

  start(configuration: Configuration, deps: GlobalDependencies)
    : BluebirdPromise<void> {
    const that = this;
    const app = Express();

    const appConfiguration = ConfigurationParser.parse(configuration);

    // by default the level of logs is info
    deps.winston.level = appConfiguration.logs_level;
    this.displayConfigurations(appConfiguration);

    return this.setup(appConfiguration, app, deps)
      .then(function () {
        return that.startServer(app, appConfiguration.port);
      });
  }

  stop() {
    this.httpServer.close();
  }
}

