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

export class Configurator {
  static configure(config: Configuration,
    app: Express.Application,
    vars: ServerVariables,
    deps: GlobalDependencies): void {
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
    app.use(function (req, res, next) {
      if (!req.session) {
        return next(new Error("No session available."));
      }
      next();
    });

    RestApi.setup(app, vars);
  }
}