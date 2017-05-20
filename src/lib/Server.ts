
import { UserConfiguration } from "./Configuration";
import { GlobalDependencies } from "../types/Dependencies";
import AuthenticationRegulator from "./AuthenticationRegulator";
import UserDataStore from "./UserDataStore";
import ConfigurationAdapter from "./ConfigurationAdapter";
import { NotifierFactory } from "./notifiers/NotifierFactory";
import TOTPValidator from "./TOTPValidator";
import TOTPGenerator from "./TOTPGenerator";
import RestApi from "./RestApi";

import * as Express from "express";
import * as BodyParser from "body-parser";
import * as Path from "path";
import * as http from "http";

import AccessController from "./access_control/AccessController";

const Ldap = require("./ldap");

export default class Server {
  private httpServer: http.Server;

  start(yaml_configuration: UserConfiguration, deps: GlobalDependencies): Promise<void> {
    const config = ConfigurationAdapter.adapt(yaml_configuration);

    const view_directory = Path.resolve(__dirname, "../views");
    const public_html_directory = Path.resolve(__dirname, "../public_html");
    const datastore_options = {
      directory: config.store_directory,
      inMemory: config.store_in_memory
    };

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
    app.set("view engine", "ejs");

    // by default the level of logs is info
    deps.winston.level = config.logs_level || "info";

    const five_minutes = 5 * 60;
    const data_store = new UserDataStore(datastore_options);
    const regulator = new AuthenticationRegulator(data_store, five_minutes);
    const notifier = NotifierFactory.build(config.notifier, deps);
    const ldap = new Ldap(deps, config.ldap);
    const accessController = new AccessController(config.access_control, deps.winston);
    const totpValidator = new TOTPValidator(deps.speakeasy);
    const totpGenerator = new TOTPGenerator(deps.speakeasy);

    app.set("logger", deps.winston);
    app.set("ldap", ldap);
    app.set("totp validator", totpValidator);
    app.set("totp generator", totpGenerator);
    app.set("u2f", deps.u2f);
    app.set("user data store", data_store);
    app.set("notifier", notifier);
    app.set("authentication regulator", regulator);
    app.set("config", config);
    app.set("access controller", accessController);

    RestApi.setup(app);

    return new Promise<void>((resolve, reject) => {
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

