
import { Winston } from "winston";
import objectPath = require("object-path");
import BluebirdPromise = require("bluebird");
import express = require("express");
import { AccessController } from "../../access_control/AccessController";
import exceptions = require("../../Exceptions");
import winston = require("winston");
import AuthenticationValidator = require("../../AuthenticationValidator");
import ErrorReplies = require("../../ErrorReplies");
import ServerVariables = require("../../ServerVariables");
import AuthenticationSession = require("../../AuthenticationSession");

function verify_filter(req: express.Request, res: express.Response): BluebirdPromise<void> {
  const logger = ServerVariables.getLogger(req.app);
  const accessController = ServerVariables.getAccessController(req.app);
  const authSession = AuthenticationSession.get(req);

  logger.debug("Verify: headers are %s", JSON.stringify(req.headers));
  authSession.redirect = "https://" + req.headers["host"] + req.headers["x-original-uri"];

  return AuthenticationValidator.validate(req)
    .then(function () {
      const username = authSession.userid;
      const groups = authSession.groups;

      const host = objectPath.get<express.Request, string>(req, "headers.host");
      const domain = host.split(":")[0];
      console.log(domain);

      const isAllowed = accessController.isDomainAllowedForUser(domain, username, groups);
      if (!isAllowed) return BluebirdPromise.reject(
        new exceptions.DomainAccessDenied("User '" + username + "' does not have access to " + domain));

      if (!authSession.first_factor || !authSession.second_factor)
        return BluebirdPromise.reject(new exceptions.AccessDeniedError("First or second factor not validated"));

      return BluebirdPromise.resolve();
    });
}

export default function (req: express.Request, res: express.Response) {
  const logger = ServerVariables.getLogger(req.app);
  verify_filter(req, res)
    .then(function () {
      res.status(204);
      res.send();
    })
    .catch(ErrorReplies.replyWithError401(res, logger));
}

