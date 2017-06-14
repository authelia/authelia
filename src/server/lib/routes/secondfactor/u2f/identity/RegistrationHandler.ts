
import BluebirdPromise = require("bluebird");
import express = require("express");
import objectPath = require("object-path");

import { IdentityValidable } from "../../../../IdentityCheckMiddleware";
import { Identity } from "../../../../../../types/Identity";
import { PRE_VALIDATION_TEMPLATE } from "../../../../IdentityCheckPreValidationTemplate";
import FirstFactorValidator = require("../../../../FirstFactorValidator");
import AuthenticationSession = require("../../../../AuthenticationSession");

const CHALLENGE = "u2f-register";
const MAIL_SUBJECT = "Register your U2F device";

const POST_VALIDATION_TEMPLATE_NAME = "u2f-register";


export default class RegistrationHandler implements IdentityValidable {
  challenge(): string {
    return CHALLENGE;
  }

  private retrieveIdentity(req: express.Request) {
    const authSession = AuthenticationSession.get(req);
    const userid = authSession.userid;
    const email = authSession.email;

    if (!(userid && email)) {
      return BluebirdPromise.reject("User ID or email is missing");
    }

    const identity = {
      email: email,
      userid: userid
    };
    return BluebirdPromise.resolve(identity);
  }

  preValidationInit(req: express.Request): BluebirdPromise<Identity> {
    const that = this;
    return FirstFactorValidator.validate(req)
      .then(function () {
        return that.retrieveIdentity(req);
      });
  }

  preValidationResponse(req: express.Request, res: express.Response) {
    res.render(PRE_VALIDATION_TEMPLATE);
  }

  postValidationInit(req: express.Request) {
    return FirstFactorValidator.validate(req);
  }

  postValidationResponse(req: express.Request, res: express.Response) {
    res.render(POST_VALIDATION_TEMPLATE_NAME);
  }

  mailSubject(): string {
    return MAIL_SUBJECT;
  }
}

