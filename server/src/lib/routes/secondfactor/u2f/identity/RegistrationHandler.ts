
import BluebirdPromise = require("bluebird");
import express = require("express");

import { IdentityValidable } from "../../../../IdentityValidable";
import { Identity } from "../../../../../../types/Identity";
import FirstFactorValidator = require("../../../../FirstFactorValidator");
import { AuthenticationSessionHandler } from "../../../../AuthenticationSessionHandler";
import { IRequestLogger } from "../../../../logging/IRequestLogger";

const CHALLENGE = "u2f-register";
const MAIL_SUBJECT = "Register your security key with Authelia";


export default class RegistrationHandler implements IdentityValidable {
  private logger: IRequestLogger;

  constructor(logger: IRequestLogger) {
    this.logger = logger;
  }

  challenge(): string {
    return CHALLENGE;
  }

  private retrieveIdentity(req: express.Request): BluebirdPromise<Identity> {
    const that = this;
    return new BluebirdPromise(function(resolve, reject) {
      const authSession = AuthenticationSessionHandler.get(req, that.logger);
      const userid = authSession.userid;
      const email = authSession.email;

      if (!(userid && email)) {
        return reject(new Error("User ID or email is missing"));
      }

      const identity = {
        email: email,
        userid: userid
      };
      return resolve(identity);
    });
  }

  preValidationInit(req: express.Request): BluebirdPromise<Identity> {
    const that = this;
    return FirstFactorValidator.validate(req, this.logger)
      .then(function () {
        return that.retrieveIdentity(req);
      });
  }

  preValidationResponse(req: express.Request, res: express.Response) {
    res.json({message: "OK"});
  }

  postValidationInit(req: express.Request) {
    return FirstFactorValidator.validate(req, this.logger);
  }

  postValidationResponse(req: express.Request, res: express.Response) {
    res.json({message: "OK"});
  }

  mailSubject(): string {
    return MAIL_SUBJECT;
  }

  destinationPath(): string {
    return "/security-key-registration";
  }
}

