import BluebirdPromise = require("bluebird");
import express = require("express");
import ErrorReplies = require("../../ErrorReplies");
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";
import UserMessages = require("../../UserMessages");
import { ServerVariables } from "../../ServerVariables";
import { AuthenticationSession } from "../../../../types/AuthenticationSession";
import { GroupsAndEmails } from "../../authentication/backends/GroupsAndEmails";
import { Level } from "../../authentication/Level";
import { Level as AuthorizationLevel } from "../../authorization/Level";
import { BelongToDomain } from "../../BelongToDomain";
import { URLDecomposer } from "../..//utils/URLDecomposer";
import { Object } from "../../../lib/authorization/Object";
import { Subject } from "../../../lib/authorization/Subject";
import AuthenticationError from "../../../lib/authentication/AuthenticationError";
import IsRedirectionSafe from "../../../lib/utils/IsRedirectionSafe";
import * as URLParse from "url-parse";
import GetHeader from "../../utils/GetHeader";

export default function (vars: ServerVariables) {
  return function (req: express.Request, res: express.Response)
    : BluebirdPromise<void> {
    const username: string = req.body.username;
    const password: string = req.body.password;
    const keepMeLoggedIn: boolean = req.body.keepMeLoggedIn;
    let authSession: AuthenticationSession;

    if (keepMeLoggedIn) {
      // Stay connected for 1 year.
      vars.logger.debug(req, "User requested to stay logged in for one year.");
      req.session.cookie.maxAge = 365 * 24 * 60 * 60 * 1000;
    }

    return BluebirdPromise.resolve()
      .then(function () {
        if (!username || !password) {
          return BluebirdPromise.reject(new Error("No username or password."));
        }
        vars.logger.info(req, "Starting authentication of user \"%s\"", username);
        authSession = AuthenticationSessionHandler.get(req, vars.logger);
        return vars.regulator.regulate(username);
      })
      .then(function () {
        vars.logger.info(req, "No regulation applied.");
        return vars.usersDatabase.checkUserPassword(username, password);
      })
      .then(function (groupsAndEmails: GroupsAndEmails) {
        vars.logger.info(req,
          "LDAP binding successful. Retrieved information about user are %s",
          JSON.stringify(groupsAndEmails));
        authSession.userid = username;
        authSession.keep_me_logged_in = keepMeLoggedIn;
        authSession.authentication_level = Level.ONE_FACTOR;

        const emails: string[] = groupsAndEmails.emails;
        const groups: string[] = groupsAndEmails.groups;

        if (emails.length > 0)
          authSession.email = emails[0];
        authSession.groups = groups;

        vars.logger.debug(req, "Mark successful authentication to regulator.");
        vars.regulator.mark(username, true);
      })
      .then(function() {
        const targetUrl = GetHeader(req, "x-target-url");

        if (!targetUrl) {
          res.status(204);
          res.send();
          return BluebirdPromise.resolve();
        }

        if (BelongToDomain(targetUrl, vars.config.session.domain)) {
          const resource = URLDecomposer.fromUrl(targetUrl);
          const resObject: Object = {
            domain: resource.domain,
            resource: resource.path,
          };

          const subject: Subject = {
            user: authSession.userid,
            groups: authSession.groups
          };

          const authorizationLevel = vars.authorizer.authorization(resObject, subject, req.ip);
          if (authorizationLevel <= AuthorizationLevel.ONE_FACTOR) {
            if (IsRedirectionSafe(vars, new URLParse(targetUrl))) {
              res.json({redirect: targetUrl});
              return BluebirdPromise.resolve();
            } else {
              res.json({error: "You're authenticated but cannot be automatically redirected to an unsafe URL."});
              return BluebirdPromise.resolve();
            }
          }
        }

        res.status(204);
        res.send();
        return BluebirdPromise.resolve();
      })
      .catch(AuthenticationError, function (err: Error) {
        vars.regulator.mark(username, false);
        return ErrorReplies.replyWithError200(req, res, vars.logger, UserMessages.AUTHENTICATION_FAILED)(err);
      })
      .catch(ErrorReplies.replyWithError200(req, res, vars.logger, UserMessages.AUTHENTICATION_FAILED));
  };
}
