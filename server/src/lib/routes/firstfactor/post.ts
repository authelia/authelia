
import Exceptions = require("../../Exceptions");
import objectPath = require("object-path");
import BluebirdPromise = require("bluebird");
import express = require("express");
import { AccessController } from "../../access_control/AccessController";
import { Regulator } from "../../regulation/Regulator";
import { GroupsAndEmails } from "../../ldap/IClient";
import Endpoint = require("../../../../../shared/api");
import ErrorReplies = require("../../ErrorReplies");
import { ServerVariablesHandler } from "../../ServerVariablesHandler";
import AuthenticationSession = require("../../AuthenticationSession");
import Constants = require("../../../../../shared/constants");
import { DomainExtractor } from "../../utils/DomainExtractor";
import UserMessages = require("../../../../../shared/UserMessages");
import { AuthenticationMethodCalculator } from "../../AuthenticationMethodCalculator";
import { ServerVariables } from "../../ServerVariables";

export default function (vars: ServerVariables) {
  return function (req: express.Request, res: express.Response)
    : BluebirdPromise<void> {
    const username: string = req.body.username;
    const password: string = req.body.password;
    let authSession: AuthenticationSession.AuthenticationSession;

    return BluebirdPromise.resolve()
      .then(function () {
        if (!username || !password) {
          return BluebirdPromise.reject(new Error("No username or password."));
        }
        vars.logger.info(req, "Starting authentication of user \"%s\"", username);
        return AuthenticationSession.get(req);
      })
      .then(function (_authSession: AuthenticationSession.AuthenticationSession) {
        authSession = _authSession;
        return vars.regulator.regulate(username);
      })
      .then(function () {
        vars.logger.info(req, "No regulation applied.");
        return vars.ldapAuthenticator.authenticate(username, password);
      })
      .then(function (groupsAndEmails: GroupsAndEmails) {
        vars.logger.info(req, "LDAP binding successful. Retrieved information about user are %s",
          JSON.stringify(groupsAndEmails));
        authSession.userid = username;
        authSession.first_factor = true;
        const redirectUrl = req.query[Constants.REDIRECT_QUERY_PARAM] !== "undefined"
          // Fuck, don't know why it is a string!
          ? req.query[Constants.REDIRECT_QUERY_PARAM]
          : undefined;

        const emails: string[] = groupsAndEmails.emails;
        const groups: string[] = groupsAndEmails.groups;
        const redirectHost: string = DomainExtractor.fromUrl(redirectUrl);
        const authMethod =
          new AuthenticationMethodCalculator(vars.config.authentication_methods)
            .compute(redirectHost);
        vars.logger.debug(req, "Authentication method for \"%s\" is \"%s\"", redirectHost, authMethod);

        if (!emails || emails.length <= 0) {
          const errMessage =
            "No emails found. The user should have at least one email address to reset password.";
          vars.logger.error(req, "%s", errMessage);
          return BluebirdPromise.reject(new Error(errMessage));
        }

        authSession.email = emails[0];
        authSession.groups = groups;

        vars.logger.debug(req, "Mark successful authentication to regulator.");
        vars.regulator.mark(username, true);

        if (authMethod == "basic_auth") {
          res.send({
            redirect: redirectUrl
          });
          vars.logger.debug(req, "Redirect to '%s'", redirectUrl);
        }
        else if (authMethod == "two_factor") {
          let newRedirectUrl = Endpoint.SECOND_FACTOR_GET;
          if (redirectUrl) {
            newRedirectUrl += "?" + Constants.REDIRECT_QUERY_PARAM + "="
              + encodeURIComponent(redirectUrl);
          }
          vars.logger.debug(req, "Redirect to '%s'", newRedirectUrl, typeof redirectUrl);
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