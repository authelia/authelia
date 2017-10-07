
import express = require("express");
import BluebirdPromise = require("bluebird");
import objectPath = require("object-path");
import exceptions = require("../../../Exceptions");
import { ServerVariablesHandler } from "../../../ServerVariablesHandler";
import AuthenticationSession = require("../../../AuthenticationSession");
import ErrorReplies = require("../../../ErrorReplies");

import Constants = require("./../constants");

export default function (req: express.Request, res: express.Response): BluebirdPromise<void> {
  const logger = ServerVariablesHandler.getLogger(req.app);
  const ldapPasswordUpdater = ServerVariablesHandler.getLdapPasswordUpdater(req.app);
  let authSession: AuthenticationSession.AuthenticationSession;
  const newPassword = objectPath.get<express.Request, string>(req, "body.password");

  return AuthenticationSession.get(req)
    .then(function (_authSession) {
      authSession = _authSession;
      logger.info(req, "User %s wants to reset his/her password.",
        authSession.identity_check.userid);
      logger.debug(req, "Challenge %s", authSession.identity_check.challenge);

      if (authSession.identity_check.challenge != Constants.CHALLENGE) {
        res.status(403);
        res.send();
        return BluebirdPromise.reject(new Error("Bad challenge."));
      }
      return ldapPasswordUpdater.updatePassword(authSession.identity_check.userid, newPassword);
    })
    .then(function () {
      logger.info(req, "Password reset for user '%s'",
        authSession.identity_check.userid);
      AuthenticationSession.reset(req);
      res.status(204);
      res.send();
      return BluebirdPromise.resolve();
    })
    .catch(ErrorReplies.replyWithError500(req, res, logger));
}
