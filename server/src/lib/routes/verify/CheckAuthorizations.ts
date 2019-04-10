import * as Util from "util";

import Exceptions = require("../../Exceptions");

import { Level as AuthorizationLevel } from "../../authorization/Level";
import { Level as AuthenticationLevel } from "../../authentication/Level";
import { IAuthorizer } from "../../authorization/IAuthorizer";

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
  authorizer: IAuthorizer,
  domain: string, resource: string,
  user: string, groups: string[], ip: string,
  authenticationLevel: AuthenticationLevel): void {

  const authorizationLevel = authorizer
    .authorization({domain, resource}, {user, groups}, ip);

  if (authorizationLevel == AuthorizationLevel.BYPASS) {
    return;
  }
  else if (user && authorizationLevel == AuthorizationLevel.DENY) {
    throw new Exceptions.NotAuthorizedError(
      Util.format("User %s is not authorized to access %s%s", (user) ? user : "unknown", domain, resource));
  }
  else if (!isAuthorized(authorizationLevel, authenticationLevel)) {
    throw new Exceptions.NotAuthenticatedError(Util.format(
      "User '%s' is not sufficiently authorized to access %s%s.", (user) ? user : "unknown", domain, resource));
  }
}