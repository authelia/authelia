
import objectPath = require("object-path");
import BluebirdPromise = require("bluebird");
import express = require("express");
import exceptions = require("../../Exceptions");
import winston = require("winston");
import AuthenticationValidator = require("../../AuthenticationValidator");
import ErrorReplies = require("../../ErrorReplies");
import { ServerVariablesHandler } from "../../ServerVariablesHandler";
import AuthenticationSession = require("../../AuthenticationSession");
import Constants = require("../../../../../shared/constants");

function verify_filter(req: express.Request, res: express.Response): BluebirdPromise<void> {
  const logger = ServerVariablesHandler.getLogger(req.app);
  const accessController = ServerVariablesHandler.getAccessController(req.app);

  return AuthenticationSession.get(req)
    .then(function (authSession) {
      logger.debug("Verify: headers are %s", JSON.stringify(req.headers));
      res.set("Redirect", encodeURIComponent("https://" + req.headers["host"] + req.headers["x-original-uri"]));

      const username = authSession.userid;
      const groups = authSession.groups;
      const onlyBasicAuth = req.query[Constants.ONLY_BASIC_AUTH_QUERY_PARAM] === "true";
      logger.debug("Verify: %s=%s", Constants.ONLY_BASIC_AUTH_QUERY_PARAM, onlyBasicAuth);

      const host = objectPath.get<express.Request, string>(req, "headers.host");
      const path = objectPath.get<express.Request, string>(req, "headers.x-original-uri");

      const domain = host.split(":")[0];
      logger.debug("Verify: domain=%s, path=%s", domain, path);
      logger.debug("Verify: user=%s, groups=%s", username, groups.join(","));

      if (!authSession.first_factor)
        return BluebirdPromise.reject(new exceptions.AccessDeniedError("First factor not validated."));

      const isAllowed = accessController.isAccessAllowed(domain, path, username, groups);
      if (!isAllowed) return BluebirdPromise.reject(
          new exceptions.DomainAccessDenied("User '" + username + "' does not have access to " + domain));

      if (!onlyBasicAuth && !authSession.second_factor)
        return BluebirdPromise.reject(new exceptions.AccessDeniedError("Second factor not validated."));

      res.setHeader("Remote-User", username);
      res.setHeader("Remote-Groups", groups.join(","));

      return BluebirdPromise.resolve();
    });
}

export default function (req: express.Request, res: express.Response): BluebirdPromise<void> {
  const logger = ServerVariablesHandler.getLogger(req.app);
  return verify_filter(req, res)
    .then(function () {
      res.status(204);
      res.send();
      return BluebirdPromise.resolve();
    })
    .catch(exceptions.DomainAccessDenied, ErrorReplies.replyWithError403(res, logger))
    .catch(ErrorReplies.replyWithError401(res, logger));
}

