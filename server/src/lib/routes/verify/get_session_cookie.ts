import Express = require("express");
import BluebirdPromise = require("bluebird");
import Util = require("util");
import ObjectPath = require("object-path");

import Exceptions = require("../../Exceptions");
import { Configuration } from "../../configuration/schema/Configuration";
import Constants = require("../../../../../shared/constants");
import { DomainExtractor } from "../../../../../shared/DomainExtractor";
import { ServerVariables } from "../../ServerVariables";
import { MethodCalculator } from "../../authentication/MethodCalculator";
import { IRequestLogger } from "../../logging/IRequestLogger";
import { AuthenticationSession }
  from "../../../../types/AuthenticationSession";
import { AuthenticationSessionHandler }
  from "../../AuthenticationSessionHandler";
import AccessControl from "./access_control";

const FIRST_FACTOR_NOT_VALIDATED_MESSAGE = "First factor not yet validated";
const SECOND_FACTOR_NOT_VALIDATED_MESSAGE = "Second factor not yet validated";

function verify_inactivity(req: Express.Request,
  authSession: AuthenticationSession,
  configuration: Configuration, logger: IRequestLogger)
  : BluebirdPromise<void> {

  // If inactivity is not specified, then inactivity timeout does not apply
  if (!configuration.session.inactivity || authSession.keep_me_logged_in) {
    return BluebirdPromise.resolve();
  }

  const lastActivityTime = authSession.last_activity_datetime;
  const currentTime = new Date().getTime();
  authSession.last_activity_datetime = currentTime;

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

export default function (req: Express.Request, res: Express.Response,
  vars: ServerVariables, authSession: AuthenticationSession)
  : BluebirdPromise<{ username: string, groups: string[] }> {
  let username: string;
  let groups: string[];
  let domain: string;
  let originalUri: string;

  return new BluebirdPromise(function (resolve, reject) {
    username = authSession.userid;
    groups = authSession.groups;

    if (!authSession.userid) {
      reject(new Exceptions.AccessDeniedError(
        Util.format("%s: %s.", FIRST_FACTOR_NOT_VALIDATED_MESSAGE,
          "userid is missing")));
      return;
    }

    const originalUrl = ObjectPath.get<Express.Request, string>(req, "headers.x-original-url");
    originalUri =
      ObjectPath.get<Express.Request, string>(req, "headers.x-original-uri");

    domain = DomainExtractor.fromUrl(originalUrl);
    const authenticationMethod =
      MethodCalculator.compute(vars.config.authentication_methods, domain);
    vars.logger.debug(req, "domain=%s, request_uri=%s, user=%s, groups=%s", domain,
      originalUri, username, groups.join(","));

    if (!authSession.first_factor)
      return reject(new Exceptions.AccessDeniedError(
        Util.format("%s: %s.", FIRST_FACTOR_NOT_VALIDATED_MESSAGE,
          "first factor is false")));

    if (authenticationMethod == "two_factor" && !authSession.second_factor)
      return reject(new Exceptions.AccessDeniedError(
        Util.format("%s: %s.", SECOND_FACTOR_NOT_VALIDATED_MESSAGE,
          "second factor is false")));

    resolve();
  })
    .then(function () {
      return AccessControl(req, vars, domain, originalUri, username, groups);
    })
    .then(function () {
      return verify_inactivity(req, authSession,
        vars.config, vars.logger);
    })
    .then(function () {
      return BluebirdPromise.resolve({
        username: authSession.userid,
        groups: authSession.groups
      });
    });
}