import BluebirdPromise = require("bluebird");

import { AccessController } from "./access_control/AccessController";
import { AppConfiguration, UserConfiguration } from "./configuration/Configuration";
import { GlobalDependencies } from "../../types/Dependencies";
import { AuthenticationRegulator } from "./AuthenticationRegulator";
import { UserDataStore } from "./storage/UserDataStore";
import { ConfigurationAdapter } from "./configuration/ConfigurationAdapter";
import { TOTPValidator } from "./TOTPValidator";
import { TOTPGenerator } from "./TOTPGenerator";
import { RestApi } from "./RestApi";
import { Client } from "./ldap/Client";
import { ServerVariablesHandler } from "./ServerVariablesHandler";
import { SessionConfigurationBuilder } from "./configuration/SessionConfigurationBuilder";

import * as Express from "express";
import * as BodyParser from "body-parser";
import * as Path from "path";
import * as http from "http";

// Constants

const TRUST_PROXY = "trust proxy";
const VIEWS = "views";
const VIEW_ENGINE = "view engine";
const PUG = "pug";


export default class Server {
  private httpServer: http.Server;

  private setupExpressApplication(config: AppConfiguration, app: Express.Application, deps: GlobalDependencies): void {
    const viewsDirectory = Path.resolve(__dirname, "../views");
    const publicHtmlDirectory = Path.resolve(__dirname, "../public_html");

    const expressSessionOptions = SessionConfigurationBuilder.build(config, deps);

    app.use(Express.static(publicHtmlDirectory));
    app.use(BodyParser.urlencoded({ extended: false }));
    app.use(BodyParser.json());
    app.use(deps.session(expressSessionOptions));

    app.set(TRUST_PROXY, 1);
    app.set(VIEWS, viewsDirectory);
    app.set(VIEW_ENGINE, PUG);

    RestApi.setup(app);
  }

  private adaptConfiguration(yamlConfiguration: UserConfiguration, deps: GlobalDependencies): AppConfiguration {
    const config = ConfigurationAdapter.adapt(yamlConfiguration);

    // by default the level of logs is info
    deps.winston.level = config.logs_level;
    console.log("Log level = ", deps.winston.level);

    deps.winston.debug("Content of YAML configuration file is %s", JSON.stringify(yamlConfiguration, undefined, 2));
    deps.winston.debug("Authelia configuration is %s", JSON.stringify(config, undefined, 2));
    return config;
  }

  private setup(config: AppConfiguration, app: Express.Application, deps: GlobalDependencies): BluebirdPromise<void> {
    this.setupExpressApplication(config, app, deps);
    return ServerVariablesHandler.initialize(app, config, deps);
  }

  private startServer(app: Express.Application, port: number) {
    return new BluebirdPromise<void>((resolve, reject) => {
      this.httpServer = app.listen(port, function (err: string) {
        console.log("Listening on %d...", port);
        resolve();
      });
    });
  }

  start(yamlConfiguration: UserConfiguration, deps: GlobalDependencies): BluebirdPromise<void> {
    const that = this;
    const app = Express();
    const config = this.adaptConfiguration(yamlConfiguration, deps);
    return this.setup(config, app, deps)
      .then(function () {
        return that.startServer(app, config.port);
      });
  }

  stop() {
    this.httpServer.close();
  }
}

