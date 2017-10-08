

import express = require("express");
import U2f = require("u2f");
import { ServerVariablesHandler } from "./ServerVariablesHandler";
import BluebirdPromise = require("bluebird");

export interface AuthenticationSession {
  userid: string;
  first_factor: boolean;
  second_factor: boolean;
  identity_check?: {
    challenge: string;
    userid: string;
  };
  register_request?: U2f.Request;
  sign_request?: U2f.Request;
  email: string;
  groups: string[];
  redirect?: string;
}

const INITIAL_AUTHENTICATION_SESSION: AuthenticationSession = {
  first_factor: false,
  second_factor: false,
  userid: undefined,
  email: undefined,
  groups: [],
  register_request: undefined,
  sign_request: undefined,
  identity_check: undefined,
  redirect: undefined
};

export function reset(req: express.Request): void {
  const logger = ServerVariablesHandler.getLogger(req.app);
  logger.debug(req, "Authentication session %s is being reset.", req.sessionID);
  req.session.auth = Object.assign({}, INITIAL_AUTHENTICATION_SESSION, {});
}

export function get(req: express.Request): BluebirdPromise<AuthenticationSession> {
  const logger = ServerVariablesHandler.getLogger(req.app);
  if (!req.session) {
    const errorMsg = "Something is wrong with session cookies. Please check Redis is running and Authelia can contact it.";
    logger.error(req, errorMsg);
    return BluebirdPromise.reject(new Error(errorMsg));
  }

  if (!req.session.auth) {
    logger.debug(req, "Authentication session %s was undefined. Resetting.", req.sessionID);
    reset(req);
  }

  return BluebirdPromise.resolve(req.session.auth);
}