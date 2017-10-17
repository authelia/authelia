
import { UserDataStore } from "../../../../storage/UserDataStore";

import objectPath = require("object-path");
import u2f_common = require("../U2FCommon");
import BluebirdPromise = require("bluebird");
import express = require("express");
import U2f = require("u2f");
import FirstFactorBlocker from "../../../FirstFactorBlocker";
import ErrorReplies = require("../../../../ErrorReplies");
import AuthenticationSessionHandler = require("../../../../AuthenticationSession");
import { AuthenticationSession } from "../../../../../../types/AuthenticationSession";
import UserMessages = require("../../../../../../../shared/UserMessages");
import { ServerVariables } from "../../../../ServerVariables";

export default function (vars: ServerVariables) {
  function handler(req: express.Request, res: express.Response): BluebirdPromise<void> {
    let authSession: AuthenticationSession;
    const appid: string = u2f_common.extract_app_id(req);

    return AuthenticationSessionHandler.get(req, vars.logger)
      .then(function (_authSession) {
        authSession = _authSession;

        if (!authSession.identity_check
          || authSession.identity_check.challenge != "u2f-register") {
          res.status(403);
          res.send();
          return BluebirdPromise.reject(new Error("Bad challenge."));
        }

        vars.logger.info(req, "Starting registration for appId '%s'", appid);

        return BluebirdPromise.resolve(vars.u2f.request(appid));
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

  return FirstFactorBlocker(handler, vars.logger);
}