import BluebirdPromise = require("bluebird");
import express = require("express");
import { U2FRegistrationDocument } from "../../../../storage/U2FRegistrationDocument";
import exceptions = require("../../../../Exceptions");
import ErrorReplies = require("../../../../ErrorReplies");
import { AuthenticationSessionHandler } from "../../../../AuthenticationSessionHandler";
import UserMessages = require("../../../../UserMessages");
import { ServerVariables } from "../../../../ServerVariables";
import { AuthenticationSession } from "../../../../../../types/AuthenticationSession";
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
      resolve();
    })
      .then(function () {
        return vars.userDataStore.retrieveU2FRegistration(authSession.userid, appid);
      })
      .then(function (doc: U2FRegistrationDocument): BluebirdPromise<void> {
        if (!doc)
          return BluebirdPromise.reject(new exceptions.AccessDeniedError("No U2F registration document found."));

        vars.logger.info(req, "Start authentication of app '%s'", appid);
        vars.logger.debug(req, "AppId = %s, keyHandle = %s", appid, JSON.stringify(doc.registration.keyHandle));

        const request = vars.u2f.request(appid, doc.registration.keyHandle);
        authSession.sign_request = request;
        res.json(request);
        return BluebirdPromise.resolve();
      })
      .catch(ErrorReplies.replyWithError200(req, res, vars.logger,
        UserMessages.OPERATION_FAILED));
  }
  return handler;
}
