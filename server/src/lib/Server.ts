import * as Bluebird from "bluebird";
import * as Express from "express";
import * as http from "http";

import { Configuration } from "./configuration/schema/Configuration";
import { GlobalDependencies } from "../../types/Dependencies";
import { ConfigurationParser } from "./configuration/ConfigurationParser";
import { GlobalLogger } from "./logging/GlobalLogger";
import { RequestLogger } from "./logging/RequestLogger";
import { ServerVariables } from "./ServerVariables";
import { ServerVariablesInitializer } from "./ServerVariablesInitializer";
import { Configurator } from "./web_server/Configurator";

import { GET_VARIABLE_KEY } from "./constants";

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
    if (displayableConfiguration.duo_api) {
      displayableConfiguration.duo_api.secret_key = STARS;
    }

    this.globalLogger.debug("User configuration is %s",
      JSON.stringify(displayableConfiguration, undefined, 2));
  }

  private setup(config: Configuration, app: Express.Application, deps: GlobalDependencies): Bluebird<void> {
    return ServerVariablesInitializer.initialize(
      config, this.globalLogger, this.requestLogger, deps)
      .then(function (vars: ServerVariables) {
        app.set(GET_VARIABLE_KEY, vars);
        return Configurator.configure(config, app, vars, deps);
      });
  }

  private startServer(app: Express.Application, port: number) {
    const that = this;
    that.globalLogger.info("Starting Authelia...");
    return new Bluebird<void>((resolve, reject) => {
      this.httpServer = app.listen(port, function (err: string) {
        that.globalLogger.info("Listening on port %d...", port);
        resolve();
      });
    });
  }

  start(configuration: Configuration, deps: GlobalDependencies)
    : Bluebird<void> {
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

