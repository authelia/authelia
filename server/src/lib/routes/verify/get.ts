
import objectPath = require("object-path");
import BluebirdPromise = require("bluebird");
import express = require("express");
import exceptions = require("../../Exceptions");
import winston = require("winston");
import ErrorReplies = require("../../ErrorReplies");
import { AppConfiguration } from "../../configuration/Configuration";
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";
import { AuthenticationSession } from "../../../../types/AuthenticationSession";
import Constants = require("../../../../../shared/constants");
import Util = require("util");
import { DomainExtractor } from "../../utils/DomainExtractor";
import { ServerVariables } from "../../ServerVariables";
import { MethodCalculator } from "../../authentication/MethodCalculator";
import { IRequestLogger } from "../../logging/IRequestLogger";

const FIRST_FACTOR_NOT_VALIDATED_MESSAGE = "First factor not yet validated";
const SECOND_FACTOR_NOT_VALIDATED_MESSAGE = "Second factor not yet validated";

const REMOTE_USER = "Remote-User";
const REMOTE_GROUPS = "Remote-Groups";

function verify_inactivity(req: express.Request,
  authSession: AuthenticationSession,
  configuration: AppConfiguration, logger: IRequestLogger)
  : BluebirdPromise<void> {

  const lastActivityTime = authSession.last_activity_datetime;
  const currentTime = new Date().getTime();
  authSession.last_activity_datetime = currentTime;

  // If inactivity is not specified, then inactivity timeout does not apply
  if (!configuration.session.inactivity) {
    return BluebirdPromise.resolve();
  }

  const inactivityPeriodMs = currentTime - lastActivityTime;
  logger.debug(req, "Inactivity period was %s s and max period was %s.",
    inactivityPeriodMs / 1000, configuration.session.inactivity / 1000);
  if (inactivityPeriodMs < configuration.session.inactivity) {
    return BluebirdPromise.resolve();
  }

  logger.debug(req, "Session has been reset after too long inactivity period.");
  AuthenticationSessionHandler.reset(req);
  return BluebirdPromise.reject(new Error("Inactivity period exceeded."));
}

function verify_filter(req: express.Request, res: express.Response,
  vars: ServerVariables): BluebirdPromise<void> {
  let authSession: AuthenticationSession;
  let username: string;
  let groups: string[];

  return new BluebirdPromise(function (resolve, reject) {
    authSession = AuthenticationSessionHandler.get(req, vars.logger);
    username = authSession.userid;
    groups = authSession.groups;

    res.set("Redirect", encodeURIComponent("https://" + req.headers["host"] +
      req.headers["x-original-uri"]));

    if (!authSession.userid) {
      reject(new exceptions.AccessDeniedError(
        Util.format("%s: %s.", FIRST_FACTOR_NOT_VALIDATED_MESSAGE, "userid is missing")));
      return;
    }

    const host = objectPath.get<express.Request, string>(req, "headers.host");
    const path = objectPath.get<express.Request, string>(req, "headers.x-original-uri");

    const domain = DomainExtractor.fromHostHeader(host);
    const authenticationMethod =
      MethodCalculator.compute(vars.config.authentication_methods, domain);
    vars.logger.debug(req, "domain=%s, path=%s, user=%s, groups=%s", domain, path,
      username, groups.join(","));

    if (!authSession.first_factor)
      return reject(new exceptions.AccessDeniedError(
        Util.format("%s: %s.", FIRST_FACTOR_NOT_VALIDATED_MESSAGE, "first factor is false")));

    if (authenticationMethod == "two_factor" && !authSession.second_factor)
      return reject(new exceptions.AccessDeniedError(
        Util.format("%s: %s.", SECOND_FACTOR_NOT_VALIDATED_MESSAGE, "second factor is false")));

    const isAllowed = vars.accessController.isAccessAllowed(domain, path, username, groups);
    if (!isAllowed) return reject(
      new exceptions.DomainAccessDenied(Util.format("User '%s' does not have access to '%s'",
        username, domain)));

    resolve();
  })
    .then(function () {
      return verify_inactivity(req, authSession,
        vars.config, vars.logger);
    })
    .then(function () {
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

