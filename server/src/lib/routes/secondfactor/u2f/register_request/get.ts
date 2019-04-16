
import BluebirdPromise = require("bluebird");
import express = require("express");
import U2f = require("u2f");
import ErrorReplies = require("../../../../ErrorReplies");
import { AuthenticationSessionHandler } from "../../../../AuthenticationSessionHandler";
import { AuthenticationSession } from "../../../../../../types/AuthenticationSession";
import UserMessages = require("../../../../UserMessages");
import { ServerVariables } from "../../../../ServerVariables";
import GetHeader from "../../../../utils/GetHeader";
import * as Constants from "../../../../constants";

export default function (vars: ServerVariables) {
  function handler(req: express.Request, res: express.Response): BluebirdPromise<void> {
    let authSession: AuthenticationSession;
    const scheme = GetHeader(req, Constants.HEADER_X_FORWARDED_PROTO);
    const host = GetHeader(req, Constants.HEADER_X_FORWARDED_HOST);
    const appid = scheme + "://" + host;

    return new BluebirdPromise(function (resolve, reject) {
      authSession = AuthenticationSessionHandler.get(req, vars.logger);
      if (!authSession.identity_check
        || authSession.identity_check.challenge != "u2f-register") {
        return reject(new Error("Bad challenge."));
      }

      vars.logger.info(req, "Starting registration for appId '%s'", appid);
      return resolve(vars.u2f.request(appid));
    })
      .then(function (registrationRequest: U2f.Request) {
        vars.logger.debug(req, "RegistrationRequest = %s", JSON.stringify(registrationRequest));
        authSession.register_request = registrationRequest;
        res.json(registrationRequest);
        return BluebirdPromise.resolve();
      })
      .catch(ErrorReplies.replyWithError200(req, res, vars.logger,
        UserMessages.OPERATION_FAILED));
  }

  return handler;
}