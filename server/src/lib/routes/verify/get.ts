import BluebirdPromise = require("bluebird");
import Express = require("express");
import Exceptions = require("../../Exceptions");
import ErrorReplies = require("../../ErrorReplies");
import { ServerVariables } from "../../ServerVariables";
import GetWithSessionCookieMethod from "./get_session_cookie";
import GetWithBasicAuthMethod from "./get_basic_auth";
import Constants = require("../../../../../shared/constants");
import { AuthenticationSessionHandler }
  from "../../AuthenticationSessionHandler";
import { AuthenticationSession }
  from "../../../../types/AuthenticationSession";
import GetHeader from "../../utils/GetHeader";

const REMOTE_USER = "Remote-User";
const REMOTE_GROUPS = "Remote-Groups";


function verifyWithSelectedMethod(req: Express.Request, res: Express.Response,
  vars: ServerVariables, authSession: AuthenticationSession | undefined)
  : () => BluebirdPromise<{ username: string, groups: string[] }> {
  return function () {
    const authorization = GetHeader(req, Constants.HEADER_PROXY_AUTHORIZATION);
    if (authorization) {
      if (authorization.startsWith("Basic ")) {
        return GetWithBasicAuthMethod(req, res, vars, authorization);
      }
      else {
        throw new Error("The authorization header should be of the form 'Basic XXXXXX'");
      }
    }
    else {
      if (authSession) {
        return GetWithSessionCookieMethod(req, res, vars, authSession);
      }
      else {
        throw new Error("No cookie detected.");
      }
    }
  };
}

function setRedirectHeader(req: Express.Request, res: Express.Response) {
  return function () {
    const originalUrl = GetHeader(req, Constants.HEADER_X_ORIGINAL_URL);
    res.set(Constants.HEADER_REDIRECT, originalUrl);
    return BluebirdPromise.resolve();
  };
}

function setUserAndGroupsHeaders(res: Express.Response) {
  return function (u: { username: string, groups: string[] }) {
    res.setHeader(REMOTE_USER, u.username);
    res.setHeader(REMOTE_GROUPS, u.groups.join(","));
    return BluebirdPromise.resolve();
  };
}

function replyWith200(res: Express.Response) {
  return function () {
    res.status(204);
    res.send();
  };
}

function getRedirectParam(req: Express.Request) {
  return req.query[Constants.REDIRECT_QUERY_PARAM] != "undefined"
    ? req.query[Constants.REDIRECT_QUERY_PARAM]
    : undefined;
}

export default function (vars: ServerVariables) {
  return function (req: Express.Request, res: Express.Response)
    : BluebirdPromise<void> {
    let authSession: AuthenticationSession | undefined;
    return new BluebirdPromise(function (resolve, reject) {
      authSession = AuthenticationSessionHandler.get(req, vars.logger);
      resolve();
    })
      .then(setRedirectHeader(req, res))
      .then(verifyWithSelectedMethod(req, res, vars, authSession))
      .then(setUserAndGroupsHeaders(res))
      .then(replyWith200(res))
      // The user is authenticated but has restricted access -> 403
      .catch(Exceptions.NotAuthorizedError,
        ErrorReplies.replyWithError403(req, res, vars.logger))
      .catch(Exceptions.NotAuthenticatedError,
        ErrorReplies.replyWithError401(req, res, vars.logger))
      // The user is not yet authenticated -> 401
      .catch((err) => {
        console.error(err);
        // This redirect parameter is used in Kubernetes to annotate the ingress with
        // the url to the authentication portal.
        const redirectUrl = getRedirectParam(req);
        if (redirectUrl) {
          ErrorReplies.redirectTo(redirectUrl, req, res, vars.logger)(err);
        }
        else {
          ErrorReplies.replyWithError401(req, res, vars.logger)(err);
        }
      });
  };
}

