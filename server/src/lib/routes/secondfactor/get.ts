
import Express = require("express");
import Endpoints = require("../../../../../shared/api");
import FirstFactorBlocker = require("../FirstFactorBlocker");
import BluebirdPromise = require("bluebird");
import AuthenticationSessionHandler = require("../../AuthenticationSession");
import { ServerVariables } from "../../ServerVariables";

const TEMPLATE_NAME = "secondfactor";

export default function (vars: ServerVariables) {
  function handler(req: Express.Request, res: Express.Response): BluebirdPromise<void> {
    return AuthenticationSessionHandler.get(req, vars.logger)
      .then(function (authSession) {
        if (authSession.first_factor && authSession.second_factor) {
          res.redirect(Endpoints.LOGGED_IN);
          return BluebirdPromise.resolve();
        }

        res.render(TEMPLATE_NAME, {
          username: authSession.userid,
          totp_identity_start_endpoint: Endpoints.SECOND_FACTOR_TOTP_IDENTITY_START_GET,
          u2f_identity_start_endpoint: Endpoints.SECOND_FACTOR_U2F_IDENTITY_START_GET
        });
        return BluebirdPromise.resolve();
      });
  }

  return FirstFactorBlocker.default(handler, vars.logger);
}