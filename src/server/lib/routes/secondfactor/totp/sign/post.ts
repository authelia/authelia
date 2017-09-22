
import exceptions = require("../../../../Exceptions");
import objectPath = require("object-path");
import express = require("express");
import { TOTPSecretDocument } from "../../../../storage/TOTPSecretDocument";
import BluebirdPromise = require("bluebird");
import FirstFactorBlocker from "../../../FirstFactorBlocker";
import Endpoints = require("../../../../../endpoints");
import redirect from "../../redirect";
import ErrorReplies = require("../../../../ErrorReplies");
import { ServerVariablesHandler } from "./../../../../ServerVariablesHandler";
import AuthenticationSession = require("../../../../AuthenticationSession");

const UNAUTHORIZED_MESSAGE = "Unauthorized access";

export default FirstFactorBlocker(handler);

export function handler(req: express.Request, res: express.Response): BluebirdPromise<void> {
  let authSession: AuthenticationSession.AuthenticationSession;
  const logger = ServerVariablesHandler.getLogger(req.app);
  const token = req.body.token;
  const totpValidator = ServerVariablesHandler.getTOTPValidator(req.app);
  const userDataStore = ServerVariablesHandler.getUserDataStore(req.app);

  return AuthenticationSession.get(req)
    .then(function (_authSession: AuthenticationSession.AuthenticationSession) {
      authSession = _authSession;
      logger.info("POST 2ndfactor totp: Initiate TOTP validation for user %s", authSession.userid);
      return userDataStore.retrieveTOTPSecret(authSession.userid);
    })
    .then(function (doc: TOTPSecretDocument) {
      logger.debug("POST 2ndfactor totp: TOTP secret is %s", JSON.stringify(doc));
      return totpValidator.validate(token, doc.secret.base32);
    })
    .then(function () {
      logger.debug("POST 2ndfactor totp: TOTP validation succeeded");
      authSession.second_factor = true;
      redirect(req, res);
      return BluebirdPromise.resolve();
    })
    .catch(exceptions.InvalidTOTPError, ErrorReplies.replyWithError401(res, logger))
    .catch(ErrorReplies.replyWithError500(res, logger));
}
