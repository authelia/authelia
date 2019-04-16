import objectPath = require("object-path");
import BluebirdPromise = require("bluebird");
import express = require("express");
import U2f = require("u2f");
import { U2FRegistration } from "../../../../../../types/U2FRegistration";
import redirect from "../../redirect";
import ErrorReplies = require("../../../../ErrorReplies");
import { ServerVariables } from "../../../../ServerVariables";
import { AuthenticationSessionHandler } from "../../../../AuthenticationSessionHandler";
import UserMessages = require("../../../../UserMessages");
import { AuthenticationSession } from "../../../../../../types/AuthenticationSession";
import GetHeader from "../../../../utils/GetHeader";
import * as Constants from "../../../../constants";


export default function (vars: ServerVariables) {
  function handler(req: express.Request, res: express.Response): BluebirdPromise<void> {
    let authSession: AuthenticationSession;
    const scheme = GetHeader(req, Constants.HEADER_X_FORWARDED_PROTO);
    const host = GetHeader(req, Constants.HEADER_X_FORWARDED_HOST);
    const appid = scheme + "://" + host;
    const registrationResponse: U2f.RegistrationData = req.body;

    return new BluebirdPromise(function (resolve, reject) {
      authSession = AuthenticationSessionHandler.get(req, vars.logger);
      const registrationRequest = authSession.register_request;

      if (!registrationRequest) {
        return reject(new Error("No registration request"));
      }

      if (!authSession.identity_check
        || authSession.identity_check.challenge != "u2f-register") {
        return reject(new Error("Bad challenge for registration request"));
      }

      vars.logger.info(req, "Finishing registration");
      vars.logger.debug(req, "RegistrationRequest = %s", JSON.stringify(registrationRequest));
      vars.logger.debug(req, "RegistrationResponse = %s", JSON.stringify(registrationResponse));

      return resolve(vars.u2f.checkRegistration(registrationRequest, registrationResponse));
    })
      .then(function (u2fResult: U2f.RegistrationResult | U2f.Error): BluebirdPromise<void> {
        if (objectPath.has(u2fResult, "errorCode"))
          return BluebirdPromise.reject(new Error("Error while registering."));

        const registrationResult: U2f.RegistrationResult = u2fResult as U2f.RegistrationResult;
        vars.logger.info(req, "Store registration and reply");
        vars.logger.debug(req, "RegistrationResult = %s", JSON.stringify(registrationResult));
        const registration: U2FRegistration = {
          keyHandle: registrationResult.keyHandle,
          publicKey: registrationResult.publicKey
        };
        return vars.userDataStore.saveU2FRegistration(authSession.userid, appid, registration);
      })
      .then(function () {
        authSession.identity_check = undefined;
        redirect(vars)(req, res);
        return BluebirdPromise.resolve();
      })
      .catch(ErrorReplies.replyWithError200(req, res, vars.logger,
        UserMessages.OPERATION_FAILED));
  }
  return handler;
}
