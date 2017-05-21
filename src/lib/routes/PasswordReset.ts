
import BluebirdPromise = require("bluebird");
import objectPath = require("object-path");
import exceptions = require("../Exceptions");
import express = require("express");
import { Identity } from "../../types/Identity";
import { IdentityValidable } from "../IdentityValidator";

const CHALLENGE = "reset-password";

class PasswordResetHandler implements IdentityValidable {
  challenge(): string {
    return CHALLENGE;
  }

  templateName(): string {
    return "reset-password";
  }

  preValidation(req: express.Request): BluebirdPromise<Identity> {
    const userid = objectPath.get(req, "body.userid");
    if (!userid) {
      return BluebirdPromise.reject(new exceptions.AccessDeniedError("No user id provided"));
    }

    const ldap = req.app.get("ldap");
    return ldap.get_emails(userid)
      .then(function (emails: string[]) {
        if (!emails && emails.length <= 0) throw new Error("No email found");

        const identity = {
          email: emails[0],
          userid: userid
        };
        return BluebirdPromise.resolve(identity);
      });
  }

  mailSubject(): string {
    return "Reset your password";
  }
}

function protect(fn: express.RequestHandler) {
  return function (req: express.Request, res: express.Response) {
    const challenge = objectPath.get(req, "session.auth_session.identity_check.challenge");
    if (challenge != CHALLENGE) {
      res.status(403);
      res.send();
      return;
    }
    fn(req, res, undefined);
  };
}

function post(req: express.Request, res: express.Response) {
  const logger = req.app.get("logger");
  const ldap = req.app.get("ldap");
  const new_password = objectPath.get(req, "body.password");
  const userid = objectPath.get(req, "session.auth_session.identity_check.userid");

  logger.info("POST reset-password: User %s wants to reset his/her password", userid);

  ldap.update_password(userid, new_password)
    .then(function () {
      logger.info("POST reset-password: Password reset for user %s", userid);
      objectPath.set(req, "session.auth_session", undefined);
      res.status(204);
      res.send();
    })
    .catch(function (err: Error) {
      logger.error("POST reset-password: Error while resetting the password of user %s. %s", userid, err);
      res.status(500);
      res.send();
    });
}

export = {
  icheck_interface: new PasswordResetHandler(),
  post: protect(post)
};
