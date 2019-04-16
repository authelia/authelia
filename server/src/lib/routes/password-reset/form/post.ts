
import express = require("express");
import BluebirdPromise = require("bluebird");
import objectPath = require("object-path");
import { AuthenticationSessionHandler } from "../../../AuthenticationSessionHandler";
import { AuthenticationSession } from "../../../../../types/AuthenticationSession";
import ErrorReplies = require("../../../ErrorReplies");
import UserMessages = require("../../../UserMessages");
import { ServerVariables } from "../../../ServerVariables";

import Constants = require("./../constants");

export default function (vars: ServerVariables) {
  return function (req: express.Request, res: express.Response): BluebirdPromise<void> {
    let authSession: AuthenticationSession;
    const newPassword = objectPath.get<express.Request, string>(req, "body.password");

    return new BluebirdPromise(function (resolve, reject) {
      authSession = AuthenticationSessionHandler.get(req, vars.logger);
      if (!authSession.identity_check) {
        reject(new Error("No identity check initiated"));
        return;
      }

      vars.logger.info(req, "User %s wants to reset his/her password.",
        authSession.identity_check.userid);
      vars.logger.debug(req, "Challenge %s", authSession.identity_check.challenge);

      if (authSession.identity_check.challenge != Constants.CHALLENGE) {
        reject(new Error("Bad challenge."));
        return;
      }
      resolve();
    })
      .then(function () {
        return vars.usersDatabase.updatePassword(authSession.identity_check.userid, newPassword);
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
