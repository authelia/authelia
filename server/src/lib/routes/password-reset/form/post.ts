
import express = require("express");
import BluebirdPromise = require("bluebird");
import objectPath = require("object-path");
import exceptions = require("../../../Exceptions");
import AuthenticationSessionHandler = require("../../../AuthenticationSession");
import { AuthenticationSession } from "../../../../../types/AuthenticationSession";
import ErrorReplies = require("../../../ErrorReplies");
import UserMessages = require("../../../../../../shared/UserMessages");
import { ServerVariables } from "../../../ServerVariables";

import Constants = require("./../constants");

export default function (vars: ServerVariables) {
  return function (req: express.Request, res: express.Response): BluebirdPromise<void> {
    let authSession: AuthenticationSession;
    const newPassword = objectPath.get<express.Request, string>(req, "body.password");

    return AuthenticationSessionHandler.get(req, vars.logger)
      .then(function (_authSession) {
        authSession = _authSession;
        vars.logger.info(req, "User %s wants to reset his/her password.",
          authSession.identity_check.userid);
        vars.logger.debug(req, "Challenge %s", authSession.identity_check.challenge);

        if (authSession.identity_check.challenge != Constants.CHALLENGE) {
          return BluebirdPromise.reject(new Error("Bad challenge."));
        }
        return vars.ldapPasswordUpdater.updatePassword(authSession.identity_check.userid, newPassword);
      })
      .then(function () {
        vars.logger.info(req, "Password reset for user '%s'",
          authSession.identity_check.userid);
        AuthenticationSessionHandler.reset(req);
        res.status(204);
        res.send();
        return BluebirdPromise.resolve();
      })
      .catch(ErrorReplies.replyWithError200(req, res, vars.logger,
        UserMessages.RESET_PASSWORD_FAILED));
  };
}
