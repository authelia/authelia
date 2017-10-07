
import objectPath = require("object-path");
import u2f_common = require("../U2FCommon");
import BluebirdPromise = require("bluebird");
import express = require("express");
import { UserDataStore } from "../../../../storage/UserDataStore";
import { U2FRegistrationDocument } from "../../../../storage/U2FRegistrationDocument";
import { Winston } from "../../../../../../types/Dependencies";
import U2f = require("u2f");
import exceptions = require("../../../../Exceptions");
import FirstFactorBlocker from "../../../FirstFactorBlocker";
import redirect from "../../redirect";
import ErrorReplies = require("../../../../ErrorReplies");
import { ServerVariablesHandler } from "../../../../ServerVariablesHandler";
import AuthenticationSession = require("../../../../AuthenticationSession");

export default FirstFactorBlocker(handler);

export function handler(req: express.Request, res: express.Response): BluebirdPromise<void> {
  const logger = ServerVariablesHandler.getLogger(req.app);
  const userDataStore = ServerVariablesHandler.getUserDataStore(req.app);
  let authSession: AuthenticationSession.AuthenticationSession;

  return AuthenticationSession.get(req)
    .then(function (_authSession: AuthenticationSession.AuthenticationSession) {
      authSession = _authSession;
      if (!authSession.sign_request) {
        const err = new Error("No sign request");
        ErrorReplies.replyWithError401(res, logger)(err);
        return BluebirdPromise.reject(err);
      }

      const userid = authSession.userid;
      const appid = u2f_common.extract_app_id(req);
      return userDataStore.retrieveU2FRegistration(userid, appid);
    })
    .then(function (doc: U2FRegistrationDocument): BluebirdPromise<U2f.SignatureResult | U2f.Error> {
      const appId = u2f_common.extract_app_id(req);
      const u2f = ServerVariablesHandler.getU2F(req.app);
      const signRequest = authSession.sign_request;
      const signData: U2f.SignatureData = req.body;
      logger.info("U2F sign: Finish authentication");
      return BluebirdPromise.resolve(u2f.checkSignature(signRequest, signData, doc.registration.publicKey));
    })
    .then(function (result: U2f.SignatureResult | U2f.Error): BluebirdPromise<void> {
      if (objectPath.has(result, "errorCode"))
        return BluebirdPromise.reject(new Error("Error while signing"));
      logger.info("U2F sign: Authentication successful");
      authSession.second_factor = true;
      redirect(req, res);
      return BluebirdPromise.resolve();
    })
    .catch(ErrorReplies.replyWithError500(res, logger));
}

