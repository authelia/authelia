import Express = require("express");
import Exceptions = require("../../Exceptions");
import ErrorReplies = require("../../ErrorReplies");
import { ServerVariables } from "../../ServerVariables";
import GetSessionCookie from "./GetSessionCookie";
import GetBasicAuth from "./GetBasicAuth";
import Constants = require("../../../../../shared/constants");
import { AuthenticationSessionHandler }
  from "../../AuthenticationSessionHandler";
import { AuthenticationSession }
  from "../../../../types/AuthenticationSession";
import HasHeader from "../..//utils/HasHeader";
import { RequestUrlGetter } from "../../utils/RequestUrlGetter";


async function verifyWithSelectedMethod(req: Express.Request, res: Express.Response,
  vars: ServerVariables, authSession: AuthenticationSession | undefined)
  : Promise<void> {
  if (HasHeader(req, Constants.HEADER_PROXY_AUTHORIZATION)) {
    await GetBasicAuth(req, res, vars);
  }
  else {
    await GetSessionCookie(req, res, vars, authSession);
  }
}

/**
 * The Redirect header is used to set the target URL in the login portal.
 *
 * @param req The request to extract X-Original-Url from.
 * @param res The response to write Redirect header to.
 */
function setRedirectHeader(req: Express.Request, res: Express.Response) {
  const originalUrl = RequestUrlGetter.getOriginalUrl(req);
  res.set(Constants.HEADER_REDIRECT, originalUrl);
}

function getRedirectParam(req: Express.Request) {
  return req.query[Constants.REDIRECT_QUERY_PARAM] != "undefined"
    ? req.query[Constants.REDIRECT_QUERY_PARAM]
    : undefined;
}

export default function (vars: ServerVariables) {
  return async function (req: Express.Request, res: Express.Response)
    : Promise<void> {
    const authSession = AuthenticationSessionHandler.get(req, vars.logger);
    setRedirectHeader(req, res);

    try {
      await verifyWithSelectedMethod(req, res, vars, authSession);
      res.status(204);
      res.send();
    } catch (err) {
      // This redirect parameter is used in Kubernetes to annotate the ingress with
      // the url to the authentication portal.
      const redirectUrl = getRedirectParam(req);
      if (redirectUrl) {
        ErrorReplies.redirectTo(redirectUrl, req, res, vars.logger)(err);
        return;
      }

      if (err instanceof Exceptions.NotAuthorizedError) {
        ErrorReplies.replyWithError403(req, res, vars.logger)(err);
      } else {
        ErrorReplies.replyWithError401(req, res, vars.logger)(err);
      }
    }
  };
}

