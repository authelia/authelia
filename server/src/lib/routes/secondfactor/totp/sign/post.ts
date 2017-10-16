
import exceptions = require("../../../../Exceptions");
import objectPath = require("object-path");
import express = require("express");
import { TOTPSecretDocument } from "../../../../storage/TOTPSecretDocument";
import BluebirdPromise = require("bluebird");
import FirstFactorBlocker from "../../../FirstFactorBlocker";
import Endpoints = require("../../../../../../../shared/api");
import redirect from "../../redirect";
import ErrorReplies = require("../../../../ErrorReplies");
import { ServerVariablesHandler } from "./../../../../ServerVariablesHandler";
import AuthenticationSession = require("../../../../AuthenticationSession");
import UserMessages = require("../../../../../../../shared/UserMessages");
import { ServerVariables } from "../../../../ServerVariables";

const UNAUTHORIZED_MESSAGE = "Unauthorized access";

export default function (vars: ServerVariables) {
  function handler(req: express.Request, res: express.Response): BluebirdPromise<void> {
    let authSession: AuthenticationSession.AuthenticationSession;
    const token = req.body.token;

    return AuthenticationSession.get(req)
      .then(function (_authSession: AuthenticationSession.AuthenticationSession) {
        authSession = _authSession;
        vars.logger.info(req, "Initiate TOTP validation for user '%s'.", authSession.userid);
        return vars.userDataStore.retrieveTOTPSecret(authSession.userid);
      })
      .then(function (doc: TOTPSecretDocument) {
        vars.logger.debug(req, "TOTP secret is %s", JSON.stringify(doc));

        if (!vars.totpHandler.validate(token, doc.secret.base32))
          return BluebirdPromise.reject(new Error("Invalid TOTP token."));

        vars.logger.debug(req, "TOTP validation succeeded.");
        authSession.second_factor = true;
        redirect(req, res);
        return BluebirdPromise.resolve();
      })
      .catch(ErrorReplies.replyWithError200(req, res, vars.logger,
        UserMessages.OPERATION_FAILED));
  }
  return FirstFactorBlocker(handler);
}
