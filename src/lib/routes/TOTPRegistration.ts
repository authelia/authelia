import objectPath = require("object-path");
import BluebirdPromise = require("bluebird");
import express = require("express");
import exceptions = require("../Exceptions");
import { Identity } from "../../types/Identity";
import { IdentityValidable } from "../IdentityValidator";

const CHALLENGE = "totp-register";
const TEMPLATE_NAME = "totp-register";


class TOTPRegistrationHandler implements IdentityValidable {
  challenge(): string {
    return CHALLENGE;
  }

  templateName(): string {
    return TEMPLATE_NAME;
  }

  preValidation(req: express.Request): BluebirdPromise<Identity> {
    const first_factor_passed = objectPath.get(req, "session.auth_session.first_factor");
    if (!first_factor_passed) {
      return BluebirdPromise.reject("Authentication required before registering TOTP secret key");
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
    return "Register your TOTP secret key";
  }
}

// Generate a secret and send it to the user
function post(req: express.Request, res: express.Response) {
  const logger = req.app.get("logger");
  const userid = objectPath.get(req, "session.auth_session.identity_check.userid");
  const challenge = objectPath.get(req, "session.auth_session.identity_check.challenge");

  if (challenge != CHALLENGE || !userid) {
    res.status(403);
    res.send();
    return;
  }

  const user_data_store = req.app.get("user data store");
  const totpGenerator = req.app.get("totp generator");
  const secret = totpGenerator.generate();

  logger.debug("POST new-totp-secret: save the TOTP secret in DB");
  user_data_store.set_totp_secret(userid, secret)
    .then(function () {
      const doc = {
        otpauth_url: secret.otpauth_url,
        base32: secret.base32,
        ascii: secret.ascii
      };
      objectPath.set(req, "session", undefined);

      res.status(200);
      res.json(doc);
    })
    .catch(function (err: Error) {
      logger.error("POST new-totp-secret: Internal error %s", err);
      res.status(500);
      res.send();
    });
}


export = {
  icheck_interface: new TOTPRegistrationHandler(),
  post: post,
};
