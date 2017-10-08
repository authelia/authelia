
import exceptions = require("../../Exceptions");
import objectPath = require("object-path");
import BluebirdPromise = require("bluebird");
import express = require("express");
import { AccessController } from "../../access_control/AccessController";
import { AuthenticationRegulator } from "../../AuthenticationRegulator";
import { GroupsAndEmails } from "../../ldap/IClient";
import Endpoint = require("../../../../../shared/api");
import ErrorReplies = require("../../ErrorReplies");
import { ServerVariablesHandler } from "../../ServerVariablesHandler";
import AuthenticationSession = require("../../AuthenticationSession");
import Constants = require("../../../../../shared/constants");

export default function (req: express.Request, res: express.Response): BluebirdPromise<void> {
  const username: string = req.body.username;
  const password: string = req.body.password;

  const logger = ServerVariablesHandler.getLogger(req.app);
  const ldap = ServerVariablesHandler.getLdapAuthenticator(req.app);
  const config = ServerVariablesHandler.getConfiguration(req.app);

  if (!username || !password) {
    const err = new Error("No username or password");
    ErrorReplies.replyWithError401(req, res, logger)(err);
    return BluebirdPromise.reject(err);
  }

  const regulator = ServerVariablesHandler.getAuthenticationRegulator(req.app);
  const accessController = ServerVariablesHandler.getAccessController(req.app);
  let authSession: AuthenticationSession.AuthenticationSession;

  logger.info(req, "Starting authentication of user \"%s\"", username);
  return AuthenticationSession.get(req)
    .then(function (_authSession: AuthenticationSession.AuthenticationSession) {
      authSession = _authSession;
      return regulator.regulate(username);
    })
    .then(function () {
      logger.info(req, "No regulation applied.");
      return ldap.authenticate(username, password);
    })
    .then(function (groupsAndEmails: GroupsAndEmails) {
      logger.info(req, "LDAP binding successful. Retrieved information about user are %s",
        JSON.stringify(groupsAndEmails));
      authSession.userid = username;
      authSession.first_factor = true;
      const redirectUrl = req.query[Constants.REDIRECT_QUERY_PARAM];
      const onlyBasicAuth = req.query[Constants.ONLY_BASIC_AUTH_QUERY_PARAM] === "true";

      const emails: string[] = groupsAndEmails.emails;
      const groups: string[] = groupsAndEmails.groups;

      if (!emails || emails.length <= 0) {
        const errMessage = "No emails found. The user should have at least one email address to reset password.";
        logger.error(req, "%s", errMessage);
        return BluebirdPromise.reject(new Error(errMessage));
      }

      authSession.email = emails[0];
      authSession.groups = groups;

      logger.debug(req, "Mark successful authentication to regulator.");
      regulator.mark(username, true);

      if (onlyBasicAuth) {
        res.send({
          redirect: redirectUrl
        });
        logger.debug(req, "Redirect to '%s'", redirectUrl);
      }
      else {
        let newRedirectUrl = Endpoint.SECOND_FACTOR_GET;
        if (redirectUrl !== "undefined") {
          newRedirectUrl += "?redirect=" + encodeURIComponent(redirectUrl);
        }
        logger.debug(req, "Redirect to '%s'", newRedirectUrl, typeof redirectUrl);
        res.send({
          redirect: newRedirectUrl
        });
      }
      return BluebirdPromise.resolve();
    })
    .catch(exceptions.LdapSearchError, ErrorReplies.replyWithError500(req, res, logger))
    .catch(exceptions.LdapBindError, function (err: Error) {
      regulator.mark(username, false);
      return ErrorReplies.replyWithError401(req, res, logger)(err);
    })
    .catch(exceptions.AuthenticationRegulationError, ErrorReplies.replyWithError403(req, res, logger))
    .catch(exceptions.DomainAccessDenied, ErrorReplies.replyWithError401(req, res, logger))
    .catch(ErrorReplies.replyWithError500(req, res, logger));
}
