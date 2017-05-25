
import objectPath = require("object-path");
import randomstring = require("randomstring");
import BluebirdPromise = require("bluebird");
import util = require("util");
import Exceptions = require("./Exceptions");
import fs = require("fs");
import ejs = require("ejs");
import UserDataStore from "./UserDataStore";
import { Winston } from "../../types/Dependencies";
import express = require("express");
import ErrorReplies = require("./ErrorReplies");
import ServerVariables = require("./ServerVariables");
import AuthenticationSession = require("./AuthenticationSession");

import Identity = require("../../types/Identity");
import { IdentityValidationRequestContent } from "./UserDataStore";

const filePath = __dirname + "/../resources/email-template.ejs";
const email_template = fs.readFileSync(filePath, "utf8");

// IdentityValidator allows user to go through a identity validation process in two steps:
// - Request an operation to be performed (password reset, registration).
// - Confirm operation with email.

export interface IdentityValidable {
  challenge(): string;
  preValidationInit(req: express.Request): BluebirdPromise<Identity.Identity>;
  postValidationInit(req: express.Request): BluebirdPromise<void>;

  preValidationResponse(req: express.Request, res: express.Response): void; // Serves a page after identity check request
  postValidationResponse(req: express.Request, res: express.Response): void; // Serves the page if identity validated
  mailSubject(): string;
}

function issue_token(userid: string, content: Object, userDataStore: UserDataStore, logger: Winston): BluebirdPromise<string> {
  const five_minutes = 4 * 60 * 1000;
  const token = randomstring.generate({ length: 64 });
  const that = this;

  logger.debug("identity_check: issue identity token %s for 5 minutes", token);
  return userDataStore.issue_identity_check_token(userid, token, content, five_minutes)
    .then(function () {
      return BluebirdPromise.resolve(token);
    });
}

function consume_token(token: string, userDataStore: UserDataStore, logger: Winston): BluebirdPromise<IdentityValidationRequestContent> {
  logger.debug("identity_check: consume token %s", token);
  return userDataStore.consume_identity_check_token(token);
}

export function register(app: express.Application, pre_validation_endpoint: string, post_validation_endpoint: string, handler: IdentityValidable) {
  app.get(pre_validation_endpoint, get_start_validation(handler, post_validation_endpoint));
  app.get(post_validation_endpoint, get_finish_validation(handler));
}

function checkIdentityToken(req: express.Request, identityToken: string): BluebirdPromise<void> {
  if (!identityToken)
    return BluebirdPromise.reject(new Exceptions.AccessDeniedError("No identity token provided"));
  return BluebirdPromise.resolve();
}

export function get_finish_validation(handler: IdentityValidable): express.RequestHandler {
  return function (req: express.Request, res: express.Response): BluebirdPromise<void> {
    const logger = ServerVariables.getLogger(req.app);
    const userDataStore = ServerVariables.getUserDataStore(req.app);

    const authSession = AuthenticationSession.get(req);
    const identityToken = objectPath.get<express.Request, string>(req, "query.identity_token");
    logger.info("GET identity_check: identity token provided is %s", identityToken);

    return checkIdentityToken(req, identityToken)
      .then(function () {
        return handler.postValidationInit(req);
      })
      .then(function () {
        return consume_token(identityToken, userDataStore, logger);
      })
      .then(function (content: IdentityValidationRequestContent) {
        authSession.identity_check = {
          challenge: handler.challenge(),
          userid: content.userid
        };
        handler.postValidationResponse(req, res);
        return BluebirdPromise.resolve();
      })
      .catch(Exceptions.FirstFactorValidationError, ErrorReplies.replyWithError401(res, logger))
      .catch(Exceptions.AccessDeniedError, ErrorReplies.replyWithError403(res, logger))
      .catch(ErrorReplies.replyWithError500(res, logger));
  };
}


export function get_start_validation(handler: IdentityValidable, postValidationEndpoint: string): express.RequestHandler {
  return function (req: express.Request, res: express.Response): BluebirdPromise<void> {
    const logger = ServerVariables.getLogger(req.app);
    const notifier = ServerVariables.getNotifier(req.app);
    const userDataStore = ServerVariables.getUserDataStore(req.app);
    let identity: Identity.Identity;
    logger.info("Identity Validation: Start identity validation");

    return handler.preValidationInit(req)
      .then(function (id: Identity.Identity) {
        logger.debug("Identity Validation: retrieved identity is %s", JSON.stringify(id));
        identity = id;
        const email_address = objectPath.get<Identity.Identity, string>(identity, "email");
        const userid = objectPath.get<Identity.Identity, string>(identity, "userid");

        if (!(email_address && userid))
          return BluebirdPromise.reject(new Exceptions.IdentityError("Missing user id or email address"));

        return issue_token(userid, undefined, userDataStore, logger);
      })
      .then(function (token: string) {
        const host = req.get("Host");
        const link_url = util.format("https://%s%s?identity_token=%s", host, postValidationEndpoint, token);
        logger.info("POST identity_check: notification sent to user %s", identity.userid);
        return notifier.notify(identity, handler.mailSubject(), link_url);
      })
      .then(function () {
        handler.preValidationResponse(req, res);
        return BluebirdPromise.resolve();
      })
      .catch(Exceptions.FirstFactorValidationError, ErrorReplies.replyWithError401(res, logger))
      .catch(Exceptions.IdentityError, ErrorReplies.replyWithError400(res, logger))
      .catch(Exceptions.AccessDeniedError, ErrorReplies.replyWithError403(res, logger))
      .catch(ErrorReplies.replyWithError500(res, logger));
  };
}
