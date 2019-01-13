import Express = require("express");
import Endpoints = require("../../../../../shared/api");
import BluebirdPromise = require("bluebird");
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";
import { ServerVariables } from "../../ServerVariables";
import ErrorReplies = require("../../ErrorReplies");

export default function (vars: ServerVariables) {
  function handler(req: Express.Request, res: Express.Response): BluebirdPromise<void> {
    return new BluebirdPromise<void>(function (resolve, reject) {
      const authSession = AuthenticationSessionHandler.get(req, vars.logger);
      res.render("already-logged-in", {
        logout_endpoint: Endpoints.LOGOUT_POST,
        username: authSession.userid,
        redirection_url: vars.config.default_redirection_url
      });
      resolve();
    })
      .catch(ErrorReplies.replyWithError401(req, res, vars.logger));
  }

  return handler;
}
