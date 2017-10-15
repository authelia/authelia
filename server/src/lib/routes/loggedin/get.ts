import Express = require("express");
import Endpoints = require("../../../../../shared/api");
import FirstFactorBlocker from "../FirstFactorBlocker";
import BluebirdPromise = require("bluebird");

export default FirstFactorBlocker(handler);

function handler(req: Express.Request, res: Express.Response): BluebirdPromise<void> {
  res.render("already-logged-in", {
    logout_endpoint: Endpoints.LOGOUT_GET
  });
  return BluebirdPromise.resolve();
}
