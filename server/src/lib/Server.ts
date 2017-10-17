import BluebirdPromise = require("bluebird");
import ObjectPath = require("object-path");

import { AccessController } from "./access_control/AccessController";
import { AppConfiguration, UserConfiguration } from "./configuration/Configuration";
import { GlobalDependencies } from "../../types/Dependencies";
import { UserDataStore } from "./storage/UserDataStore";
import { ConfigurationParser } from "./configuration/ConfigurationParser";
import { RestApi } from "./RestApi";
import { ServerVariablesHandler, ServerVariablesInitializer } from "./ServerVariablesHandler";
import { SessionConfigurationBuilder } from "./configuration/SessionConfigurationBuilder";
import { GlobalLogger } from "./logging/GlobalLogger";
import { RequestLogger } from "./logging/RequestLogger";
import { ServerVariables } from "./ServerVariables";

import * as Express from "express";
import * as BodyParser from "body-parser";
import * as Path from "path";
import * as http from "http";

const addRequestId = require("express-request-id")();

// Constants
const TRUST_PROXY = "trust proxy";
const X_POWERED_BY = "x-powered-by";
const VIEWS = "views";
const VIEW_ENGINE = "view engine";
const PUG = "pug";

function clone(obj: any) {
  return JSON.parse(JSON.stringify(obj));
}

export default class Server {
  private httpServer: http.Server;
  private globalLogger: GlobalLogger;
  private requestLogger: RequestLogger;
  private serverVariables: ServerVariables;

  constructor(deps: GlobalDependencies) {
    this.globalLogger = new GlobalLogger(deps.winston);
    this.requestLogger = new RequestLogger(deps.winston);
  }

  private setupExpressApplication(config: AppConfiguration,
    app: Express.Application,
    deps: GlobalDependencies): void {
    const viewsDirectory = Path.resolve(__dirname, "../views");
    const publicHtmlDirectory = Path.resolve(__dirname, "../public_html");

    const expressSessionOptions = SessionConfigurationBuilder.build(config, deps);

    app.use(Express.static(publicHtmlDirectory));
    app.use(BodyParser.urlencoded({ extended: false }));
    app.use(BodyParser.json());
    app.use(deps.session(expressSessionOptions));
    app.use(addRequestId);
    app.disable(X_POWERED_BY);
    app.enable(TRUST_PROXY);

    app.set(VIEWS, viewsDirectory);
    app.set(VIEW_ENGINE, PUG);

    RestApi.setup(app, this.serverVariables);
  }

  private displayConfigurations(userConfiguration: UserConfiguration,
    appConfiguration: AppConfiguration) {
    const displayableUserConfiguration = clone(userConfiguration);
    const displayableAppConfiguration = clone(appConfiguration);
    const STARS = "*****";

    displayableUserConfiguration.ldap.password = STARS;
    displayableUserConfiguration.session.secret = STARS;
    if (displayableUserConfiguration.notifier && displayableUserConfiguration.notifier.gmail)
      displayableUserConfiguration.notifier.gmail.password = STARS;
    if (displayableUserConfiguration.notifier && displayableUserConfiguration.notifier.smtp)
      displayableUserConfiguration.notifier.smtp.password = STARS;

    displayableAppConfiguration.ldap.password = STARS;
    displayableAppConfiguration.session.secret = STARS;
    if (displayableAppConfiguration.notifier && displayableAppConfiguration.notifier.gmail)
      displayableAppConfiguration.notifier.gmail.password = STARS;
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
        that.serverVariables = vars;
        that.setupExpressApplication(config, app, deps);
        ServerVariablesHandler.setup(app, vars);
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

