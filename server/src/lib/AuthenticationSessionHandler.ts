

import express = require("express");
import { AuthenticationSession } from "../../types/AuthenticationSession";
import { IRequestLogger } from "./logging/IRequestLogger";
import { Level } from "./authentication/Level";

const INITIAL_AUTHENTICATION_SESSION: AuthenticationSession = {
  keep_me_logged_in: false,
  authentication_level: Level.NOT_AUTHENTICATED,
  last_activity_datetime: undefined,
  userid: undefined,
  email: undefined,
  groups: [],
  register_request: undefined,
  sign_request: undefined,
  identity_check: undefined,
  redirect: undefined
};

export class AuthenticationSessionHandler {
  static reset(req: express.Request): void {
    req.session.auth = Object.assign({}, INITIAL_AUTHENTICATION_SESSION, {});

    // Initialize last activity with current time
    req.session.auth.last_activity_datetime = new Date().getTime();
  }

  static get(req: express.Request, logger: IRequestLogger): AuthenticationSession {
    if (!req.session) {
      const errorMsg = "Something is wrong with session cookies. Please check Redis is running and Authelia can connect to it.";
      logger.error(req, errorMsg);
      throw new Error(errorMsg);
    }

    if (!req.session.auth) {
      logger.debug(req, "Session %s has no authentication information. Its internal id is: %s its current cookie is: %s",
          req.sessionID, req.session.id, JSON.stringify(req.session.cookie));
      logger.debug(req, "Authentication session %s was undefined. Resetting..." +
        " If it's unexpected, make sure you are visiting the expected domain.", req.sessionID);
      AuthenticationSessionHandler.reset(req);
    }

    return req.session.auth;
  }
}