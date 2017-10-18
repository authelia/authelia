import express = require("express");
import BluebirdPromise = require("bluebird");
import objectPath = require("object-path");

import exceptions = require("../../../Exceptions");
import { Identity } from "../../../../../types/Identity";
import { IdentityValidable } from "../../../IdentityCheckMiddleware";
import { PRE_VALIDATION_TEMPLATE } from "../../../IdentityCheckPreValidationTemplate";
import Constants = require("../constants");
import { IRequestLogger } from "../../../logging/IRequestLogger";
import { IEmailsRetriever } from "../../../ldap/IEmailsRetriever";

export const TEMPLATE_NAME = "password-reset-form";

export default class PasswordResetHandler implements IdentityValidable {
  private logger: IRequestLogger;
  private emailsRetriever: IEmailsRetriever;

  constructor(logger: IRequestLogger, emailsRetriever: IEmailsRetriever) {
    this.logger = logger;
    this.emailsRetriever = emailsRetriever;
  }

  challenge(): string {
    return Constants.CHALLENGE;
  }

  preValidationInit(req: express.Request): BluebirdPromise<Identity> {
    const userid: string = objectPath.get<express.Request, string>(req, "query.userid");

    this.logger.debug(req, "User '%s' requested a password reset", userid);
    if (!userid)
      return BluebirdPromise.reject(new exceptions.AccessDeniedError("No user id provided"));

    return this.emailsRetriever.retrieve(userid)
      .then(function (emails: string[]) {
        if (!emails && emails.length <= 0) throw new Error("No email found");
        const identity = {
          email: emails[0],
          userid: userid
        };
        return BluebirdPromise.resolve(identity);
      });
  }

  preValidationResponse(req: express.Request, res: express.Response) {
    res.render(PRE_VALIDATION_TEMPLATE);
  }

  postValidationInit(req: express.Request) {
    return BluebirdPromise.resolve();
  }

  postValidationResponse(req: express.Request, res: express.Response) {
    res.render(TEMPLATE_NAME);
  }

  mailSubject(): string {
    return "Reset your password";
  }
}