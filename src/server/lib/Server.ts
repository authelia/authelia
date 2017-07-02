
import { AccessController } from "./access_control/AccessController";
import { UserConfiguration } from "./../../types/Configuration";
import { GlobalDependencies } from "../../types/Dependencies";
import { AuthenticationRegulator } from "./AuthenticationRegulator";
import UserDataStore from "./UserDataStore";
import ConfigurationAdapter from "./ConfigurationAdapter";
import {  TOTPValidator } from "./TOTPValidator";
import { TOTPGenerator } from "./TOTPGenerator";
import RestApi from "./RestApi";
import { LdapClient } from "./LdapClient";
import BluebirdPromise = require("bluebird");
import ServerVariables = require("./ServerVariables");
import SessionConfigurationBuilder from "./SessionConfigurationBuilder";

import * as Express from "express";
import * as BodyParser from "body-parser";
import * as Path from "path";
import * as http from "http";

export default class Server {
  private httpServer: http.Server;

  start(yamlConfiguration: UserConfiguration, deps: GlobalDependencies): BluebirdPromise<void> {
    const config = ConfigurationAdapter.adapt(yamlConfiguration);

    const viewsDirectory = Path.resolve(__dirname, "../views");
    const publicHtmlDirectory = Path.resolve(__dirname, "../public_html");

    const app = Express();
    app.use(Express.static(publicHtmlDirectory));
    app.use(BodyParser.urlencoded({ extended: false }));
    app.use(BodyParser.json());

    app.set("trust proxy", 1); // trust first proxy

    const sessionOptions = SessionConfigurationBuilder.build(config, deps);
    app.use(deps.session(sessionOptions));

    app.set("views", viewsDirectory);
    app.set("view engine", "pug");

    // by default the level of logs is info
    deps.winston.level = config.logs_level;
    console.log("Log level = ", deps.winston.level);
    deps.winston.debug("Authelia configuration is %s", JSON.stringify(config, undefined, 2));

    ServerVariables.fill(app, config, deps);

    RestApi.setup(app);

    return new BluebirdPromise<void>((resolve, reject) => {
      this.httpServer = app.listen(config.port, function (err: string) {
        console.log("Listening on %d...", config.port);
        resolve();
      });
    });
  }

  stop() {
    this.httpServer.close();
  }
}

