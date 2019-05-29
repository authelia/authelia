import Express = require("express");
import Exceptions = require("../../Exceptions");
import ErrorReplies = require("../../ErrorReplies");
import { ServerVariables } from "../../ServerVariables";
import GetSessionCookie from "./GetSessionCookie";
import GetBasicAuth from "./GetBasicAuth";
import Constants = require("../../constants");
import { AuthenticationSessionHandler }
  from "../../AuthenticationSessionHandler";
import { AuthenticationSession }
  from "../../../../types/AuthenticationSession";
import HasHeader from "../..//utils/HasHeader";
import RequestUrlGetter from "../../utils/RequestUrlGetter";


async function verifyWithSelectedMethod(req: Express.Request, res: Express.Response,
  vars: ServerVariables, authSession: AuthenticationSession | undefined)
  : Promise<void> {
  if (HasHeader(req, Constants.HEADER_PROXY_AUTHORIZATION)) {
    vars.logger.debug(req, "Got PROXY_AUTHORIZATION header checking basic auth");
    await GetBasicAuth(req, res, vars);
  } else {
    vars.logger.debug(req, "Checking session cookie");
    await GetSessionCookie(req, res, vars, authSession);
  }
}

function getRedirectParam(req: Express.Request) {
  return req.query[Constants.REDIRECT_QUERY_PARAM] != "undefined"
    ? req.query[Constants.REDIRECT_QUERY_PARAM]
    : undefined;
}

async function unsafeGet(vars: ServerVariables, req: Express.Request, res: Express.Response) {
  const authSession = AuthenticationSessionHandler.get(req, vars.logger);
  try {
    await verifyWithSelectedMethod(req, res, vars, authSession);
    res.status(204);
    res.send();
  } catch (err) {
    // Kubernetes ingress controller and Traefik use the rd parameter of the verify
    // endpoint to provide the URL of the login portal. The target URL of the user
    // is computed from X-Fowarded-* headers or X-Original-Url.
    let redirectUrl = getRedirectParam(req);
    const originalUrl = RequestUrlGetter.getOriginalUrl(req);
    if (redirectUrl && originalUrl) {
      redirectUrl = redirectUrl + `?${Constants.REDIRECT_QUERY_PARAM}=` + originalUrl;
      ErrorReplies.redirectTo(redirectUrl, req, res, vars.logger)(err);
      return;
    }

    // Reply with an error.
    vars.logger.error(req, "Got an error state when processing verify. Error was: %s", err.toString());
    throw err;
  }
}

export default function (vars: ServerVariables) {
  return async function (req: Express.Request, res: Express.Response)
    : Promise<void> {
    try {
      await unsafeGet(vars, req, res);
    } catch (err) {
      if (err instanceof Exceptions.NotAuthorizedError) {
        ErrorReplies.replyWithError403(req, res, vars.logger)(err);
      } else {
        ErrorReplies.replyWithError401(req, res, vars.logger)(err);
      }
    }
  };
}

