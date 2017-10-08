
import express = require("express");
import objectPath = require("object-path");
import winston = require("winston");
import Endpoints = require("../../../../../shared/api");
import AuthenticationValidator = require("../../AuthenticationValidator");
import { ServerVariablesHandler } from "../../ServerVariablesHandler";
import BluebirdPromise = require("bluebird");
import AuthenticationSession = require("../../AuthenticationSession");
import Constants = require("../../../../../shared/constants");
import Util = require("util");

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
    res.redirect(Util.format("%s?redirect=%s", Endpoints.SECOND_FACTOR_GET,
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

export default function (req: express.Request, res: express.Response): BluebirdPromise<void> {
  return AuthenticationSession.get(req)
    .then(function (authSession) {
      if (authSession.first_factor) {
        if (authSession.second_factor)
          redirectToService(req, res);
        else
          redirectToSecondFactorPage(req, res);
        return BluebirdPromise.resolve();
      }

      renderFirstFactor(res);
      return BluebirdPromise.resolve();
    });
}