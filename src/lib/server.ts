
import { authelia } from "../types/authelia";
import * as Express from "express";
import * as BodyParser from "body-parser";
import * as Path from "path";
import { AuthenticationRegulator } from "./AuthenticationRegulator";

const UserDataStore = require("./user_data_store");
const Notifier = require("./notifier");
const setup_endpoints = require("./setup_endpoints");
const config_adapter = require("./config_adapter");
const Ldap = require("./ldap");
const AccessControl = require("./access_control");

export function run(yaml_configuration: authelia.Configuration, deps: authelia.GlobalDependencies, fn?: () => undefined) {
  const config = config_adapter(yaml_configuration);

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
    secret: config.session_secret,
    resave: false,
    saveUninitialized: true,
    cookie: {
      secure: false,
      maxAge: config.session_max_age,
      domain: config.session_domain
    },
  }));

  app.set("views", view_directory);
  app.set("view engine", "ejs");

  // by default the level of logs is info
  deps.winston.level = config.logs_level || "info";

  const five_minutes = 5 * 60;
  const data_store = new UserDataStore(deps.nedb, datastore_options);
  const regulator = new AuthenticationRegulator(data_store, five_minutes);
  const notifier = new Notifier(config.notifier, deps);
  const ldap = new Ldap(deps, config.ldap);
  const access_control = AccessControl(deps.winston, config.access_control);

  app.set("logger", deps.winston);
  app.set("ldap", ldap);
  app.set("totp engine", deps.speakeasy);
  app.set("u2f", deps.u2f);
  app.set("user data store", data_store);
  app.set("notifier", notifier);
  app.set("authentication regulator", regulator);
  app.set("config", config);
  app.set("access control", access_control);
  setup_endpoints(app);

  return app.listen(config.port, function(err: string) {
    console.log("Listening on %d...", config.port);
    if (fn) fn();
  });
}
