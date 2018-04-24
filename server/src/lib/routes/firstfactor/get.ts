
import express = require("express");
import objectPath = require("object-path");
import winston = require("winston");
import Endpoints = require("../../../../../shared/api");
import BluebirdPromise = require("bluebird");
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";
import Constants = require("../../../../../shared/constants");
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
      encodeURIComponent(redirectUrl)));
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

export default function (vars: ServerVariables) {
  return function (req: express.Request, res: express.Response): BluebirdPromise<void> {
    return new BluebirdPromise(function (resolve, reject) {
      const authSession = AuthenticationSessionHandler.get(req, vars.logger);
      if (authSession.first_factor) {
        if (authSession.second_factor)
          redirectToService(req, res);
        else
          redirectToSecondFactorPage(req, res);
        resolve();
        return;
      }
      renderFirstFactor(res);
      resolve();
    });
  };
}