
import express = require("express");
import objectPath = require("object-path");
import winston = require("winston");
import Endpoints = require("../../../endpoints");
import AuthenticationValidator = require("../../AuthenticationValidator");
import { ServerVariablesHandler } from "../../ServerVariablesHandler";
import BluebirdPromise = require("bluebird");

export default function (req: express.Request, res: express.Response): BluebirdPromise<void> {
  const logger = ServerVariablesHandler.getLogger(req.app);

  logger.debug("First factor: headers are %s", JSON.stringify(req.headers));

  res.render("firstfactor", {
    first_factor_post_endpoint: Endpoints.FIRST_FACTOR_POST,
    reset_password_request_endpoint: Endpoints.RESET_PASSWORD_REQUEST_GET
  });
  return BluebirdPromise.resolve();
}