
import { UserDataStore } from "../../../../storage/UserDataStore";

import objectPath = require("object-path");
import u2f_common = require("../U2FCommon");
import BluebirdPromise = require("bluebird");
import express = require("express");
import U2f = require("u2f");
import { U2FRegistration } from "../../../../../../types/U2FRegistration";
import FirstFactorBlocker from "../../../FirstFactorBlocker";
import redirect from "../../redirect";
import ErrorReplies = require("../../../../ErrorReplies");
import { ServerVariablesHandler } from "../../../../ServerVariablesHandler";
import AuthenticationSession = require("../../../../AuthenticationSession");


export default FirstFactorBlocker(handler);


function handler(req: express.Request, res: express.Response): BluebirdPromise<void> {
  let authSession: AuthenticationSession.AuthenticationSession;
  const userDataStore = ServerVariablesHandler.getUserDataStore(req.app);
  const u2f = ServerVariablesHandler.getU2F(req.app);
  const appid = u2f_common.extract_app_id(req);
  const logger = ServerVariablesHandler.getLogger(req.app);
  const registrationResponse: U2f.RegistrationData = req.body;

  return AuthenticationSession.get(req)
    .then(function (_authSession: AuthenticationSession.AuthenticationSession) {
      authSession = _authSession;
      const registrationRequest = authSession.register_request;

      if (!registrationRequest) {
        res.status(403);
        res.send();
        return BluebirdPromise.reject(new Error("No registration request"));
      }

      if (!authSession.identity_check
        || authSession.identity_check.challenge != "u2f-register") {
        res.status(403);
        res.send();
        return BluebirdPromise.reject(new Error("Bad challenge for registration request"));
      }

      logger.info(req, "Finishing registration");
      logger.debug(req, "RegistrationRequest = %s", JSON.stringify(registrationRequest));
      logger.debug(req, "RegistrationResponse = %s", JSON.stringify(registrationResponse));

      return BluebirdPromise.resolve(u2f.checkRegistration(registrationRequest, registrationResponse));
    })
    .then(function (u2fResult: U2f.RegistrationResult | U2f.Error): BluebirdPromise<void> {
      if (objectPath.has(u2fResult, "errorCode"))
        return BluebirdPromise.reject(new Error("Error while registering."));

      const registrationResult: U2f.RegistrationResult = u2fResult as U2f.RegistrationResult;
      logger.info(req, "Store registration and reply");
      logger.debug(req, "RegistrationResult = %s", JSON.stringify(registrationResult));
      const registration: U2FRegistration = {
        keyHandle: registrationResult.keyHandle,
        publicKey: registrationResult.publicKey
      };
      return userDataStore.saveU2FRegistration(authSession.userid, appid, registration);
    })
    .then(function () {
      authSession.identity_check = undefined;
      redirect(req, res);
      return BluebirdPromise.resolve();
    })
    .catch(ErrorReplies.replyWithError500(req, res, logger));
}
