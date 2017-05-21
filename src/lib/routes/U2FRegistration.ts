
import objectPath = require("object-path");
import BluebirdPromise = require("bluebird");
import express = require("express");

import { IdentityValidable } from "../IdentityValidator";
import { Identity } from "../../types/Identity";

const CHALLENGE = "u2f-register";
const TEMPLATE_NAME = "u2f-register";
const MAIL_SUBJECT = "Register your U2F device";


class U2FRegistrationHandler implements IdentityValidable {
  challenge(): string {
    return CHALLENGE;
  }

  templateName(): string {
    return TEMPLATE_NAME;
  }

  preValidation(req: express.Request): BluebirdPromise<Identity> {
    const first_factor_passed = objectPath.get(req, "session.auth_session.first_factor");
    if (!first_factor_passed) {
      return BluebirdPromise.reject("Authentication required before issuing a u2f registration request");
    }

    const userid = objectPath.get<express.Request, string>(req, "session.auth_session.userid");
    const email = objectPath.get<express.Request, string>(req, "session.auth_session.email");

    if (!(userid && email)) {
      return BluebirdPromise.reject("User ID or email is missing");
    }

    const identity = {
      email: email,
      userid: userid
    };
    return BluebirdPromise.resolve(identity);
  }

  mailSubject(): string {
    return MAIL_SUBJECT;
  }
}

export = {
  icheck_interface: new U2FRegistrationHandler(),
};

