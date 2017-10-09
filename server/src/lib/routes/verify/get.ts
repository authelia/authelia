
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
import Util = require("util");
import { DomainExtractor } from "../../utils/DomainExtractor";

const FIRST_FACTOR_NOT_VALIDATED_MESSAGE = "First factor not yet validated";
const SECOND_FACTOR_NOT_VALIDATED_MESSAGE = "Second factor not yet validated";

function verify_filter(req: express.Request, res: express.Response): BluebirdPromise<void> {
  const logger = ServerVariablesHandler.getLogger(req.app);
  const accessController = ServerVariablesHandler.getAccessController(req.app);
  const authenticationMethodsCalculator = ServerVariablesHandler.getAuthenticationMethodCalculator(req.app);

  return AuthenticationSession.get(req)
    .then(function (authSession) {
      res.set("Redirect", encodeURIComponent("https://" + req.headers["host"] +
        req.headers["x-original-uri"]));

      const username = authSession.userid;
      const groups = authSession.groups;
      if (!authSession.userid)
        return BluebirdPromise.reject(
          new exceptions.AccessDeniedError(FIRST_FACTOR_NOT_VALIDATED_MESSAGE));

      const host = objectPath.get<express.Request, string>(req, "headers.host");
      const path = objectPath.get<express.Request, string>(req, "headers.x-original-uri");

      const domain = DomainExtractor.fromHostHeader(host);
      const authenticationMethod = authenticationMethodsCalculator.compute(domain);
      logger.debug(req, "domain=%s, path=%s, user=%s, groups=%s", domain, path,
        username, groups.join(","));

      if (!authSession.first_factor)
        return BluebirdPromise.reject(
          new exceptions.AccessDeniedError(FIRST_FACTOR_NOT_VALIDATED_MESSAGE));

      const isAllowed = accessController.isAccessAllowed(domain, path, username, groups);
      if (!isAllowed) return BluebirdPromise.reject(
        new exceptions.DomainAccessDenied(Util.format("User '%s' does not have access to '%'",
          username, domain)));

      if (authenticationMethod == "two_factor" && !authSession.second_factor)
        return BluebirdPromise.reject(
          new exceptions.AccessDeniedError(SECOND_FACTOR_NOT_VALIDATED_MESSAGE));

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
    .catch(exceptions.DomainAccessDenied, ErrorReplies.replyWithError403(req, res, logger))
    .catch(ErrorReplies.replyWithError401(req, res, logger));
}

