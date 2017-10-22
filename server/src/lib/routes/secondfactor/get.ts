
import Express = require("express");
import Endpoints = require("../../../../../shared/api");
import BluebirdPromise = require("bluebird");
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";
import { ServerVariables } from "../../ServerVariables";
import { MethodCalculator } from "../../authentication/MethodCalculator";

const TEMPLATE_NAME = "secondfactor";

export default function (vars: ServerVariables) {
  function handler(req: Express.Request, res: Express.Response)
    : BluebirdPromise<void> {

    return new BluebirdPromise(function (resolve, reject) {
      const isSingleFactorMode: boolean = MethodCalculator.isSingleFactorOnlyMode(
        vars.config.authentication_methods);
      const authSession = AuthenticationSessionHandler.get(req, vars.logger);
      if (isSingleFactorMode
        || (authSession.first_factor && authSession.second_factor)) {
        res.redirect(Endpoints.LOGGED_IN);
        resolve();
        return;
      }

      res.render(TEMPLATE_NAME, {
        username: authSession.userid,
        totp_identity_start_endpoint:
        Endpoints.SECOND_FACTOR_TOTP_IDENTITY_START_GET,
        u2f_identity_start_endpoint:
        Endpoints.SECOND_FACTOR_U2F_IDENTITY_START_GET
      });
      resolve();
    });
  }
  return handler;
}
