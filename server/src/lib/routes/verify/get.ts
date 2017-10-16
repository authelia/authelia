
import objectPath = require("object-path");
import BluebirdPromise = require("bluebird");
import express = require("express");
import exceptions = require("../../Exceptions");
import winston = require("winston");
import AuthenticationValidator = require("../../AuthenticationValidator");
import ErrorReplies = require("../../ErrorReplies");
import { AppConfiguration } from "../../configuration/Configuration";
import AuthenticationSession = require("../../AuthenticationSession");
import Constants = require("../../../../../shared/constants");
import Util = require("util");
import { DomainExtractor } from "../../utils/DomainExtractor";
import { ServerVariables } from "../../ServerVariables";
import { AuthenticationMethodCalculator } from "../../AuthenticationMethodCalculator";

const FIRST_FACTOR_NOT_VALIDATED_MESSAGE = "First factor not yet validated";
const SECOND_FACTOR_NOT_VALIDATED_MESSAGE = "Second factor not yet validated";

const REMOTE_USER = "Remote-User";
const REMOTE_GROUPS = "Remote-Groups";

function verify_filter(req: express.Request, res: express.Response,
  vars: ServerVariables): BluebirdPromise<void> {

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
      const authenticationMethod =
        new AuthenticationMethodCalculator(vars.config.authentication_methods)
          .compute(domain);
      vars.logger.debug(req, "domain=%s, path=%s, user=%s, groups=%s", domain, path,
        username, groups.join(","));

      if (!authSession.first_factor)
        return BluebirdPromise.reject(
          new exceptions.AccessDeniedError(FIRST_FACTOR_NOT_VALIDATED_MESSAGE));

      const isAllowed = vars.accessController.isAccessAllowed(domain, path, username, groups);
      if (!isAllowed) return BluebirdPromise.reject(
        new exceptions.DomainAccessDenied(Util.format("User '%s' does not have access to '%s'",
          username, domain)));

      if (authenticationMethod == "two_factor" && !authSession.second_factor)
        return BluebirdPromise.reject(
          new exceptions.AccessDeniedError(SECOND_FACTOR_NOT_VALIDATED_MESSAGE));

      res.setHeader(REMOTE_USER, username);
      res.setHeader(REMOTE_GROUPS, groups.join(","));

      return BluebirdPromise.resolve();
    });
}

export default function (vars: ServerVariables) {
  return function (req: express.Request, res: express.Response)
    : BluebirdPromise<void> {
    return verify_filter(req, res, vars)
      .then(function () {
        res.status(204);
        res.send();
        return BluebirdPromise.resolve();
      })
      // The user is authenticated but has restricted access -> 403
      .catch(exceptions.DomainAccessDenied, ErrorReplies
        .replyWithError403(req, res, vars.logger))
      // The user is not yet authenticated -> 401
      .catch(ErrorReplies.replyWithError401(req, res, vars.logger));
  };
}

