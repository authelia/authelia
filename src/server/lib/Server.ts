
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

import * as Express from "express";
import * as BodyParser from "body-parser";
import * as Path from "path";
import * as http from "http";

export default class Server {
  private httpServer: http.Server;

  start(yaml_configuration: UserConfiguration, deps: GlobalDependencies): BluebirdPromise<void> {
    const config = ConfigurationAdapter.adapt(yaml_configuration);

    const view_directory = Path.resolve(__dirname, "../views");
    const public_html_directory = Path.resolve(__dirname, "../public_html");

    const app = Express();
    app.use(Express.static(public_html_directory));
    app.use(BodyParser.urlencoded({ extended: false }));
    app.use(BodyParser.json());

    app.set("trust proxy", 1); // trust first proxy

    app.use(deps.session({
      secret: config.session.secret,
      resave: false,
      saveUninitialized: true,
      cookie: {
        secure: false,
        maxAge: config.session.expiration,
        domain: config.session.domain
      },
    }));

    app.set("views", view_directory);
    app.set("view engine", "pug");

    // by default the level of logs is info
    deps.winston.level = config.logs_level;
    console.log("Log level = ", deps.winston.level);

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

