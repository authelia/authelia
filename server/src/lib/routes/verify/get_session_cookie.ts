import Express = require("express");
import BluebirdPromise = require("bluebird");
import Util = require("util");
import ObjectPath = require("object-path");

import Exceptions = require("../../Exceptions");
import { Configuration } from "../../configuration/schema/Configuration";
import { ServerVariables } from "../../ServerVariables";
import { IRequestLogger } from "../../logging/IRequestLogger";
import { AuthenticationSession }
  from "../../../../types/AuthenticationSession";
import { AuthenticationSessionHandler }
  from "../../AuthenticationSessionHandler";
import AccessControl from "./access_control";
import { URLDecomposer } from "../../utils/URLDecomposer";

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

  return BluebirdPromise.resolve()
    .then(() => {
    const username = authSession.userid;
    const groups = authSession.groups;

    if (!authSession.userid) {
      return BluebirdPromise.reject(new Exceptions.AccessDeniedError(
        "userid is missing"));
    }

    const originalUrl = ObjectPath.get<Express.Request, string>(
      req, "headers.x-original-url");

    const d = URLDecomposer.fromUrl(originalUrl);
    vars.logger.debug(req, "domain=%s, path=%s, user=%s, groups=%s", d.domain,
      d.path, username, groups.join(","));
    return AccessControl(req, vars, d.domain, d.path, username, groups,
      authSession.authentication_level);
  })
    .then(() => {
      return verify_inactivity(req, authSession,
        vars.config, vars.logger);
    })
    .then(() => {
      return BluebirdPromise.resolve({
        username: authSession.userid,
        groups: authSession.groups
      });
    });
}