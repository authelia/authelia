
import exceptions = require("../../../../Exceptions");
import objectPath = require("object-path");
import express = require("express");
import { TOTPSecretDocument } from "../../../../UserDataStore";
import BluebirdPromise = require("bluebird");
import FirstFactorBlocker from "../../../FirstFactorBlocker";
import Endpoints = require("../../../../../endpoints");
import redirect from "../../redirect";
import ErrorReplies = require("../../../../ErrorReplies");
import ServerVariables = require("./../../../../ServerVariables");
import AuthenticationSession = require("../../../../AuthenticationSession");

const UNAUTHORIZED_MESSAGE = "Unauthorized access";

export default FirstFactorBlocker(handler);

export function handler(req: express.Request, res: express.Response): BluebirdPromise<void> {
  const logger = ServerVariables.getLogger(req.app);
  const authSession = AuthenticationSession.get(req);
  const userid = authSession.userid;
  logger.info("POST 2ndfactor totp: Initiate TOTP validation for user %s", userid);

  const token = req.body.token;
  const totpValidator = ServerVariables.getTOTPValidator(req.app);
  const userDataStore = ServerVariables.getUserDataStore(req.app);

  logger.debug("POST 2ndfactor totp: Fetching secret for user %s", userid);
  return userDataStore.get_totp_secret(userid)
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
