
import objectPath = require("object-path");
import u2f_common = require("../U2FCommon");
import BluebirdPromise = require("bluebird");
import express = require("express");
import { U2FRegistrationDocument } from "../../../../storage/U2FRegistrationDocument";
import U2f = require("u2f");
import redirect from "../../redirect";
import ErrorReplies = require("../../../../ErrorReplies");
import { ServerVariables } from "../../../../ServerVariables";
import { AuthenticationSessionHandler } from "../../../../AuthenticationSessionHandler";
import UserMessages = require("../../../../../../../shared/UserMessages");
import { AuthenticationSession } from "../../../../../../types/AuthenticationSession";
import { Level } from "../../../../authentication/Level";

export default function (vars: ServerVariables) {
  function handler(req: express.Request, res: express.Response): BluebirdPromise<void> {
    let authSession: AuthenticationSession;
    const appId = u2f_common.extract_app_id(req);

    return new BluebirdPromise(function (resolve, reject) {
      authSession = AuthenticationSessionHandler.get(req, vars.logger);
      if (!authSession.sign_request) {
        const err = new Error("No sign request");
        ErrorReplies.replyWithError401(req, res, vars.logger)(err);
        return reject(err);
      }
      resolve();
    })
      .then(function () {
        const userid = authSession.userid;
        return vars.userDataStore.retrieveU2FRegistration(userid, appId);
      })
      .then(function (doc: U2FRegistrationDocument): BluebirdPromise<U2f.SignatureResult | U2f.Error> {
        const signRequest = authSession.sign_request;
        const signData: U2f.SignatureData = req.body;
        vars.logger.info(req, "Finish authentication");
        return BluebirdPromise.resolve(vars.u2f.checkSignature(signRequest, signData, doc.registration.publicKey));
      })
      .then(function (result: U2f.SignatureResult | U2f.Error): BluebirdPromise<void> {
        if (objectPath.has(result, "errorCode"))
          return BluebirdPromise.reject(new Error("Error while signing"));
        vars.logger.info(req, "Successful authentication");
        authSession.authentication_level = Level.TWO_FACTOR;
        redirect(vars)(req, res);
        return BluebirdPromise.resolve();
      })
      .catch(ErrorReplies.replyWithError200(req, res, vars.logger,
        UserMessages.AUTHENTICATION_U2F_FAILED));
  }

  return handler;
}

