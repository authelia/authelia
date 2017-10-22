
import objectPath = require("object-path");
import U2f = require("u2f");
import u2f_common = require("../../../secondfactor/u2f/U2FCommon");
import BluebirdPromise = require("bluebird");
import express = require("express");
import { UserDataStore } from "../../../../storage/UserDataStore";
import { U2FRegistrationDocument } from "../../../../storage/U2FRegistrationDocument";
import { Winston } from "../../../../../../types/Dependencies";
import exceptions = require("../../../../Exceptions");
import { SignMessage } from "../../../../../../../shared/SignMessage";
import ErrorReplies = require("../../../../ErrorReplies");
import { AuthenticationSessionHandler } from "../../../../AuthenticationSessionHandler";
import UserMessages = require("../../../../../../../shared/UserMessages");
import { ServerVariables } from "../../../../ServerVariables";
import { AuthenticationSession } from "../../../../../../types/AuthenticationSession";

export default function (vars: ServerVariables) {
  function handler(req: express.Request, res: express.Response): BluebirdPromise<void> {
    let authSession: AuthenticationSession;
    const appId = u2f_common.extract_app_id(req);

    return new BluebirdPromise(function (resolve, reject) {
      authSession = AuthenticationSessionHandler.get(req, vars.logger);
      resolve();
    })
      .then(function () {
        return vars.userDataStore.retrieveU2FRegistration(authSession.userid, appId);
      })
      .then(function (doc: U2FRegistrationDocument): BluebirdPromise<SignMessage> {
        if (!doc)
          return BluebirdPromise.reject(new exceptions.AccessDeniedError("No U2F registration found"));

        const appId: string = u2f_common.extract_app_id(req);
        vars.logger.info(req, "Start authentication of app '%s'", appId);
        vars.logger.debug(req, "AppId = %s, keyHandle = %s", appId, JSON.stringify(doc.registration.keyHandle));

        const request = vars.u2f.request(appId, doc.registration.keyHandle);
        const authenticationMessage: SignMessage = {
          request: request,
          keyHandle: doc.registration.keyHandle
        };
        return BluebirdPromise.resolve(authenticationMessage);
      })
      .then(function (authenticationMessage: SignMessage) {
        vars.logger.info(req, "Store authentication request and reply");
        vars.logger.debug(req, "AuthenticationRequest = %s", authenticationMessage);
        authSession.sign_request = authenticationMessage.request;
        res.json(authenticationMessage);
        return BluebirdPromise.resolve();
      })
      .catch(ErrorReplies.replyWithError200(req, res, vars.logger,
        UserMessages.OPERATION_FAILED));
  }
  return handler;
}
