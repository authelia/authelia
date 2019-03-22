import * as Express from "express";
import { AuthenticationSession } from "AuthenticationSession";
import { Configuration } from "../../configuration/schema/Configuration";
import { IRequestLogger } from "../../logging/IRequestLogger";
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";

export default function(req: Express.Request,
  authSession: AuthenticationSession,
  configuration: Configuration, logger: IRequestLogger): void {

  // If inactivity is not specified, then inactivity timeout does not apply
  if (!configuration.session.inactivity || authSession.keep_me_logged_in) {
    return;
  }

  const lastActivityTime = authSession.last_activity_datetime;
  const currentTime = new Date().getTime();
  authSession.last_activity_datetime = currentTime;

  const inactivityPeriodMs = currentTime - lastActivityTime;
  logger.debug(req, "Inactivity period was %s sec and max period was %s sec.",
    inactivityPeriodMs / 1000, configuration.session.inactivity / 1000);

  if (inactivityPeriodMs < configuration.session.inactivity) {
    return;
  }

  logger.debug(req, "Session has been reset after too long inactivity period.");
  AuthenticationSessionHandler.reset(req);
  throw new Error("Inactivity period exceeded.");
}
