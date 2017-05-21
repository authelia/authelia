
import objectPath = require("object-path");
import randomstring = require("randomstring");
import BluebirdPromise = require("bluebird");
import util = require("util");
import exceptions = require("./Exceptions");
import fs = require("fs");
import ejs = require("ejs");
import UserDataStore from "./UserDataStore";
import { ILogger } from "../types/ILogger";
import express = require("express");

import Identity = require("../types/Identity");
import { IdentityValidationRequestContent } from "./UserDataStore";

const filePath = __dirname + "/../resources/email-template.ejs";
const email_template = fs.readFileSync(filePath, "utf8");


// IdentityValidator allows user to go through a identity validation process in two steps:
// - Request an operation to be performed (password reset, registration).
// - Confirm operation with email.

export interface IdentityValidable {
  challenge(): string;
  templateName(): string;
  preValidation(req: express.Request): BluebirdPromise<Identity.Identity>;
  mailSubject(): string;
}

export class IdentityValidator {
  private userDataStore: UserDataStore;
  private logger: ILogger;

  constructor(userDataStore: UserDataStore, logger: ILogger) {
    this.userDataStore = userDataStore;
    this.logger = logger;
  }


  static setup(app: express.Application, endpoint: string, handler: IdentityValidable, userDataStore: UserDataStore, logger: ILogger) {
    const identityValidator = new IdentityValidator(userDataStore, logger);
    app.get(endpoint, identityValidator.identity_check_get(endpoint, handler));
    app.post(endpoint, identityValidator.identity_check_post(endpoint, handler));
  }


  private issue_token(userid: string, content: Object): BluebirdPromise<string> {
    const five_minutes = 4 * 60 * 1000;
    const token = randomstring.generate({ length: 64 });
    const that = this;

    this.logger.debug("identity_check: issue identity token %s for 5 minutes", token);
    return this.userDataStore.issue_identity_check_token(userid, token, content, five_minutes)
      .then(function () {
        return BluebirdPromise.resolve(token);
      });
  }

  private consume_token(token: string): BluebirdPromise<IdentityValidationRequestContent> {
    this.logger.debug("identity_check: consume token %s", token);
    return this.userDataStore.consume_identity_check_token(token);
  }

  private identity_check_get(endpoint: string, handler: IdentityValidable): express.RequestHandler {
    const that = this;
    return function (req: express.Request, res: express.Response) {
      const logger = req.app.get("logger");
      const identity_token = objectPath.get<express.Request, string>(req, "query.identity_token");
      logger.info("GET identity_check: identity token provided is %s", identity_token);

      if (!identity_token) {
        res.status(403);
        res.send();
        return;
      }

      that.consume_token(identity_token)
        .then(function (content: IdentityValidationRequestContent) {
          objectPath.set(req, "session.auth_session.identity_check", {});
          req.session.auth_session.identity_check.challenge = handler.challenge();
          req.session.auth_session.identity_check.userid = content.userid;
          res.render(handler.templateName());
        }, function (err: Error) {
          logger.error("GET identity_check: Error while consuming token %s", err);
          throw new exceptions.AccessDeniedError("Access denied");
        })
        .catch(exceptions.AccessDeniedError, function (err: Error) {
          logger.error("GET identity_check: Access Denied %s", err);
          res.status(403);
          res.send();
        })
        .catch(function (err: Error) {
          logger.error("GET identity_check: Internal error %s", err);
          res.status(500);
          res.send();
        });
    };
  }


  private identity_check_post(endpoint: string, handler: IdentityValidable): express.RequestHandler {
    const that = this;
    return function (req: express.Request, res: express.Response) {
      const logger = req.app.get("logger");
      const notifier = req.app.get("notifier");
      let identity: Identity.Identity;

      handler.preValidation(req)
        .then(function (id: Identity.Identity) {
          identity = id;
          const email_address = objectPath.get<Identity.Identity, string>(identity, "email");
          const userid = objectPath.get<Identity.Identity, string>(identity, "userid");

          if (!(email_address && userid)) {
            throw new exceptions.IdentityError("Missing user id or email address");
          }

          return that.issue_token(userid, undefined);
        }, function (err: Error) {
          throw new exceptions.AccessDeniedError(err.message);
        })
        .then(function (token: string) {
          const redirect_url = objectPath.get<express.Request, string>(req, "body.redirect");
          const original_url = util.format("https://%s%s", req.headers.host, req.headers["x-original-uri"]);
          let link_url = util.format("%s?identity_token=%s", original_url, token);
          if (redirect_url) {
            link_url = util.format("%s&redirect=%s", link_url, redirect_url);
          }

          logger.info("POST identity_check: notify to %s", identity.userid);
          return notifier.notify(identity, handler.mailSubject(), link_url);
        })
        .then(function () {
          res.status(204);
          res.send();
        })
        .catch(exceptions.IdentityError, function (err: Error) {
          logger.error("POST identity_check: %s", err);
          res.status(400);
          res.send();
        })
        .catch(exceptions.AccessDeniedError, function (err: Error) {
          logger.error("POST identity_check: %s", err);
          res.status(403);
          res.send();
        })
        .catch(function (err: Error) {
          logger.error("POST identity_check: Error %s", err);
          res.status(500);
          res.send();
        });
    };
  }
}
