
import Express = require("express");
import Endpoints = require("../../../../../shared/api");
import BluebirdPromise = require("bluebird");
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";
import { ServerVariables } from "../../ServerVariables";

const TEMPLATE_NAME = "secondfactor";

export default function (vars: ServerVariables) {
  function handler(req: Express.Request, res: Express.Response)
    : BluebirdPromise<void> {

    return new BluebirdPromise(function (resolve, reject) {
      const authSession = AuthenticationSessionHandler.get(req, vars.logger);

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
