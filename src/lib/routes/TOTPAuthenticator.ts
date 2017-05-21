
import exceptions = require("../Exceptions");
import objectPath = require("object-path");
import express = require("express");
import { TOTPSecretDocument } from "../UserDataStore";
import BluebirdPromise = require("bluebird");

const UNAUTHORIZED_MESSAGE = "Unauthorized access";

export = function(req: express.Request, res: express.Response) {
  const logger = req.app.get("logger");
  const userid = objectPath.get(req, "session.auth_session.userid");
  logger.info("POST 2ndfactor totp: Initiate TOTP validation for user %s", userid);

  if (!userid) {
    logger.error("POST 2ndfactor totp: No user id in the session");
    res.status(403);
    res.send();
    return;
  }

  const token = req.body.token;
  const totpValidator = req.app.get("totp validator");
  const userDataStore = req.app.get("user data store");

  logger.debug("POST 2ndfactor totp: Fetching secret for user %s", userid);
  userDataStore.get_totp_secret(userid)
    .then(function (doc: TOTPSecretDocument) {
      logger.debug("POST 2ndfactor totp: TOTP secret is %s", JSON.stringify(doc));
      return totpValidator.validate(token, doc.secret.base32);
    })
    .then(function () {
      logger.debug("POST 2ndfactor totp: TOTP validation succeeded");
      objectPath.set(req, "session.auth_session.second_factor", true);
      res.status(204);
      res.send();
    })
    .catch(exceptions.InvalidTOTPError, function (err: Error) {
      logger.error("POST 2ndfactor totp: Invalid TOTP token %s", err.message);
      res.status(401);
      res.send("Invalid TOTP token");
    })
    .catch(function (err: Error) {
      console.log(err.stack);
      logger.error("POST 2ndfactor totp: Internal error %s", err.message);
      res.status(500);
      res.send("Internal error");
    });
};
