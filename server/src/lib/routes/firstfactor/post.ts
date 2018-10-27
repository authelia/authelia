
import Exceptions = require("../../Exceptions");
import objectPath = require("object-path");
import BluebirdPromise = require("bluebird");
import express = require("express");
import { AccessController } from "../../access_control/AccessController";
import { Regulator } from "../../regulation/Regulator";
import Endpoint = require("../../../../../shared/api");
import ErrorReplies = require("../../ErrorReplies");
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";
import Constants = require("../../../../../shared/constants");
import { DomainExtractor } from "../../../../../shared/DomainExtractor";
import UserMessages = require("../../../../../shared/UserMessages");
import { MethodCalculator } from "../../authentication/MethodCalculator";
import { ServerVariables } from "../../ServerVariables";
import { AuthenticationSession } from "../../../../types/AuthenticationSession";
import { GroupsAndEmails } from "../../authentication/backends/GroupsAndEmails";

export default function (vars: ServerVariables) {
  return function (req: express.Request, res: express.Response)
    : BluebirdPromise<void> {
    const username: string = req.body.username;
    const password: string = req.body.password;
    const keepMeLoggedIn: boolean = req.body.keepMeLoggedIn &&
      req.body.keepMeLoggedIn === "true";
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
        authSession.first_factor = true;
        const redirectUrl: string = req.query[Constants.REDIRECT_QUERY_PARAM] !== "undefined"
          // Fuck, don't know why it is a string!
          ? req.query[Constants.REDIRECT_QUERY_PARAM]
          : undefined;

        const emails: string[] = groupsAndEmails.emails;
        const groups: string[] = groupsAndEmails.groups;

        const domain = DomainExtractor.fromUrl(redirectUrl);
        const redirectHost = (domain) ? domain : "";
        const authMethod = MethodCalculator.compute(
          vars.config.authentication_methods, redirectHost);
        vars.logger.debug(req, "Authentication method for \"%s\" is \"%s\"",
          redirectHost, authMethod);

        if (emails.length > 0)
          authSession.email = emails[0];
        authSession.groups = groups;

        vars.logger.debug(req, "Mark successful authentication to regulator.");
        vars.regulator.mark(username, true);

        if (authMethod == "single_factor") {
          let newRedirectionUrl = redirectUrl;
          if (!newRedirectionUrl)
            newRedirectionUrl = Endpoint.LOGGED_IN;
          res.send({
            redirect: newRedirectionUrl
          });
          vars.logger.debug(req, "Redirect to '%s'", redirectUrl);
        }
        else if (authMethod == "two_factor") {
          let newRedirectUrl = Endpoint.SECOND_FACTOR_GET;
          if (redirectUrl) {
            newRedirectUrl += "?" + Constants.REDIRECT_QUERY_PARAM + "="
              + redirectUrl;
          }
          vars.logger.debug(req, "Redirect to '%s'", newRedirectUrl);
          res.send({
            redirect: newRedirectUrl
          });
        }
        else {
          return BluebirdPromise.reject(new Error("Unknown authentication method for this domain."));
        }
        return BluebirdPromise.resolve();
      })
      .catch(Exceptions.LdapBindError, function (err: Error) {
        vars.regulator.mark(username, false);
        return ErrorReplies.replyWithError200(req, res, vars.logger, UserMessages.OPERATION_FAILED)(err);
      })
      .catch(ErrorReplies.replyWithError200(req, res, vars.logger, UserMessages.OPERATION_FAILED));
  };
}