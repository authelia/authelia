import express = require("express");
import BluebirdPromise = require("bluebird");
import objectPath = require("object-path");

import exceptions = require("../../../Exceptions");
import { Identity } from "../../../../../types/Identity";
import { IdentityValidable } from "../../../IdentityValidable";
import Constants = require("../constants");
import { IRequestLogger } from "../../../logging/IRequestLogger";
import { IUsersDatabase } from "../../../authentication/backends/IUsersDatabase";

export default class PasswordResetHandler implements IdentityValidable {
  private logger: IRequestLogger;
  private usersDatabase: IUsersDatabase;

  constructor(logger: IRequestLogger, usersDatabase: IUsersDatabase) {
    this.logger = logger;
    this.usersDatabase = usersDatabase;
  }

  challenge(): string {
    return Constants.CHALLENGE;
  }

  preValidationInit(req: express.Request): BluebirdPromise<Identity> {
    const that = this;
    const userid: string =
      objectPath.get<express.Request, string>(req, "body.username");
    return BluebirdPromise.resolve()
      .then(function () {
        that.logger.debug(req, "User '%s' requested a password reset", userid);
        if (!userid) {
          return BluebirdPromise.reject(
            new exceptions.AccessDeniedError("No user id provided"));
        }
        return that.usersDatabase.getEmails(userid);
      })
      .then(function (emails: string[]) {
        if (!emails && emails.length <= 0) throw new Error("No email found");
        const identity = {
          email: emails[0],
          userid: userid
        };
        return BluebirdPromise.resolve(identity);
      })
      .catch(function (err: Error) {
        return BluebirdPromise.reject(new exceptions.IdentityError(err.message));
      });
  }

  preValidationResponse(req: express.Request, res: express.Response) {
    res.status(204);
    res.send();
  }

  postValidationInit(req: express.Request) {
    return BluebirdPromise.resolve();
  }

  postValidationResponse(req: express.Request, res: express.Response) {
    res.status(204);
    res.send();
  }

  mailSubject(): string {
    return "Reset your password";
  }

  destinationPath(): string {
    return "/reset-password";
  }
}