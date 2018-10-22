import Express = require("express");
import BluebirdPromise = require("bluebird");
import Util = require("util");

import Exceptions = require("../../Exceptions");

import { Level as AuthorizationLevel } from "../../authorization/Level";
import { Level as AuthenticationLevel } from "../../authentication/Level";
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";
import { ServerVariables } from "../../ServerVariables";

function isAuthorized(
  authorization: AuthorizationLevel,
  authentication: AuthenticationLevel): boolean {

  if (authorization == AuthorizationLevel.BYPASS) {
    return true;
  } else if (authorization == AuthorizationLevel.ONE_FACTOR &&
    authentication >= AuthenticationLevel.ONE_FACTOR) {
    return true;
  } else if (authorization == AuthorizationLevel.TWO_FACTOR &&
    authentication >= AuthenticationLevel.TWO_FACTOR) {
    return true;
  }
  return false;
}

export default function (
  req: Express.Request,
  vars: ServerVariables,
  domain: string, path: string,
  username: string, groups: string[],
  authenticationLevel: AuthenticationLevel) {

  return new BluebirdPromise(function (resolve, reject) {
    const authorizationLevel = vars.authorizer
      .authorization(domain, path, username, groups);

    if (!isAuthorized(authorizationLevel, authenticationLevel)) {
      if (authorizationLevel == AuthorizationLevel.DENY) {
        reject(new Exceptions.NotAuthorizedError(
          Util.format("User %s is unauthorized to access %s%s", username, domain, path)));
        return;
      }
      reject(new Exceptions.NotAuthenticatedError(Util.format(
        "User '%s' is not sufficiently authenticated.", username, domain, path)));
      return;
    }
    resolve();
  });
}