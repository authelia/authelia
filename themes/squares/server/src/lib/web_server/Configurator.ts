import { Configuration } from "../configuration/schema/Configuration";
import { GlobalDependencies } from "../../../types/Dependencies";
import { SessionConfigurationBuilder } from
  "../configuration/SessionConfigurationBuilder";
import Path = require("path");
import Express = require("express");
import * as BodyParser from "body-parser";
import { RestApi } from "./RestApi";
import { WithHeadersLogged } from "./middlewares/WithHeadersLogged";
import { ServerVariables } from "../ServerVariables";
import Helmet = require("helmet");

const addRequestId = require("express-request-id")();

// Constants
const TRUST_PROXY = "trust proxy";
const X_POWERED_BY = "x-powered-by";
const VIEWS = "views";
const VIEW_ENGINE = "view engine";
const PUG = "pug";

export class Configurator {
  static configure(config: Configuration,
    app: Express.Application,
    vars: ServerVariables,
    deps: GlobalDependencies): void {
    const viewsDirectory = Path.resolve(__dirname, "../../views");
    const publicHtmlDirectory = Path.resolve(__dirname, "../../public_html");

    const expressSessionOptions = SessionConfigurationBuilder.build(config, deps);

    app.use(Express.static(publicHtmlDirectory));
    app.use(BodyParser.urlencoded({ extended: false }));
    app.use(BodyParser.json());
    app.use(deps.session(expressSessionOptions));
    app.use(addRequestId);
    app.use(WithHeadersLogged.middleware(vars.logger));
    app.disable(X_POWERED_BY);
    app.enable(TRUST_PROXY);
    app.use(Helmet());

    app.set(VIEWS, viewsDirectory);
    app.set(VIEW_ENGINE, PUG);

    RestApi.setup(app, vars);
  }
}