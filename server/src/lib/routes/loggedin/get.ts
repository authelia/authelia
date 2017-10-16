import Express = require("express");
import Endpoints = require("../../../../../shared/api");
import FirstFactorBlocker from "../FirstFactorBlocker";
import BluebirdPromise = require("bluebird");
import AuthenticationSession = require("../../AuthenticationSession");

export default FirstFactorBlocker(handler);

function handler(req: Express.Request, res: Express.Response): BluebirdPromise<void> {
  return AuthenticationSession.get(req)
    .then(function (authSession) {
      res.render("already-logged-in", {
        logout_endpoint: Endpoints.LOGOUT_GET,
        username: authSession.userid
      });
    });
}
