
import express = require("express");
import objectPath = require("object-path");
import Endpoints = require("../../../../../shared/api");
import BluebirdPromise = require("bluebird");
import { AuthenticationSession } from "../../../../types/AuthenticationSession";
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";
import { WhitelistValue } from "../../authentication/whitelist/WhitelistHandler";
import Constants = require("../../../../../shared/constants");
import Endpoint = require("../../../../../shared/api");
import Util = require("util");
import { ServerVariables } from "../../ServerVariables";

function getRedirectParam(req: express.Request) {
  return req.query[Constants.REDIRECT_QUERY_PARAM] != "undefined"
    ? req.query[Constants.REDIRECT_QUERY_PARAM]
    : undefined;
}

function redirectToSecondFactorPage(req: express.Request, res: express.Response) {
  const redirectUrl = getRedirectParam(req);
  if (!redirectUrl)
    res.redirect(Endpoints.SECOND_FACTOR_GET);
  else
    res.redirect(Util.format("%s?%s=%s", Endpoints.SECOND_FACTOR_GET,
      Constants.REDIRECT_QUERY_PARAM,
      redirectUrl));
}

function redirectToService(req: express.Request, res: express.Response) {
  const redirectUrl = getRedirectParam(req);
  if (!redirectUrl)
    res.redirect(Endpoints.LOGGED_IN);
  else
    res.redirect(redirectUrl);
}

function renderFirstFactor(res: express.Response) {
  res.render("firstfactor", {
    first_factor_post_endpoint: Endpoints.FIRST_FACTOR_POST,
    reset_password_request_endpoint: Endpoints.RESET_PASSWORD_REQUEST_GET
  });
}

function redirect(req: express.Request, res: express.Response, authSession: AuthenticationSession) {
  if (authSession.first_factor) {
    if (authSession.second_factor)
      redirectToService(req, res);
    else
      redirectToSecondFactorPage(req, res);
    return;
  }
  renderFirstFactor(res);
  return;
}

export default function (vars: ServerVariables) {
  return function (req: express.Request, res: express.Response): BluebirdPromise<void> {
    return new BluebirdPromise(function (resolve, reject) {
      const authSession = AuthenticationSessionHandler.get(req, vars.logger);
      // If cookie has userid and is whitelisted, user probably doesn't have whitelist access control
      // or is deliberately navigating to the auth page
      if (authSession.userid && authSession.whitelisted > WhitelistValue.NOT_WHITELISTED) {
        return redirect(req, res, authSession);
      }

      // Check for whitelisted user on request and handle auto-login
      vars.whitelistHandler.isWhitelisted(req.ip, vars.usersDatabase)
        .then((user) => {
          if (user) {
            vars.logger.info(req, "Whitelisted IP matched to user \"%s\"", user);
            vars.whitelistHandler.loginWhitelistUser(user, req, vars)
              .then(() => {
                redirectToService(req, res);
                return resolve();
              });
          } else {
            redirect(req, res, authSession);
          }
        })
        .catch(() => {
          renderFirstFactor(res);
          resolve();
        });
    });
  };
}
