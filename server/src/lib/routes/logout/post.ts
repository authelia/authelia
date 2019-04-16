
import express = require("express");
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";
import Constants = require("../../constants");
import { ServerVariables } from "../../ServerVariables";

function getRedirectParam(req: express.Request) {
  return req.query[Constants.REDIRECT_QUERY_PARAM] != "undefined"
    ? req.query[Constants.REDIRECT_QUERY_PARAM]
    : undefined;
}

export default function (vars: ServerVariables) {
  return function(req: express.Request, res: express.Response) {
    const redirect_param = getRedirectParam(req);
    const redirect_url = redirect_param || "/";
    AuthenticationSessionHandler.reset(req);
    res.redirect(redirect_url);
  };
}