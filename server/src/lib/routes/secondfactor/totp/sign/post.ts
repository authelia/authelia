
import exceptions = require("../../../../Exceptions");
import objectPath = require("object-path");
import express = require("express");
import { TOTPSecretDocument } from "../../../../storage/TOTPSecretDocument";
import BluebirdPromise = require("bluebird");
import FirstFactorBlocker from "../../../FirstFactorBlocker";
import Endpoints = require("../../../../../../../shared/api");
import redirect from "../../redirect";
import ErrorReplies = require("../../../../ErrorReplies");
import { ServerVariablesHandler } from "./../../../../ServerVariablesHandler";
import AuthenticationSession = require("../../../../AuthenticationSession");
import UserMessages = require("../../../../../../../shared/UserMessages");

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
      logger.info(req, "Initiate TOTP validation for user '%s'.", authSession.userid);
      return userDataStore.retrieveTOTPSecret(authSession.userid);
    })
    .then(function (doc: TOTPSecretDocument) {
      logger.debug(req, "TOTP secret is %s", JSON.stringify(doc));
      return totpValidator.validate(token, doc.secret.base32);
    })
    .then(function () {
      logger.debug(req, "TOTP validation succeeded.");
      authSession.second_factor = true;
      redirect(req, res);
      return BluebirdPromise.resolve();
    })
    .catch(ErrorReplies.replyWithError200(req, res, logger,
      UserMessages.OPERATION_FAILED));
}
