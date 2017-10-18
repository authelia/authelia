import Express = require("express");
import Endpoints = require("../../../../../shared/api");
import FirstFactorBlocker from "../FirstFactorBlocker";
import BluebirdPromise = require("bluebird");
import AuthenticationSessionHandler = require("../../AuthenticationSession");
import { ServerVariables } from "../../ServerVariables";

export default function (vars: ServerVariables) {
  function handler(req: Express.Request, res: Express.Response): BluebirdPromise<void> {
    return AuthenticationSessionHandler.get(req, vars.logger)
      .then(function (authSession) {
        res.render("already-logged-in", {
          logout_endpoint: Endpoints.LOGOUT_GET,
          username: authSession.userid,
          redirection_url: vars.config.default_redirection_url
        });
      });
  }

  return FirstFactorBlocker(handler, vars.logger);
}
