
import exceptions = require("../Exceptions");
import objectPath = require("object-path");
import BluebirdPromise = require("bluebird");
import express = require("express");
import AccessController from "../access_control/AccessController";
import AuthenticationRegulator from "../AuthenticationRegulator";
import { LdapClient } from "../LdapClient";

export = function (req: express.Request, res: express.Response) {
  const username: string = req.body.username;
  const password: string = req.body.password;
  if (!username || !password) {
    res.status(401);
    res.send();
    return;
  }

  const logger = req.app.get("logger");
  const ldap: LdapClient = req.app.get("ldap");
  const config = req.app.get("config");
  const regulator: AuthenticationRegulator = req.app.get("authentication regulator");
  const accessController: AccessController = req.app.get("access controller");

  logger.info("1st factor: Starting authentication of user \"%s\"", username);
  logger.debug("1st factor: Start bind operation against LDAP");
  logger.debug("1st factor: username=%s", username);

  regulator.regulate(username)
    .then(function () {
      return ldap.bind(username, password);
    })
    .then(function () {
      objectPath.set(req, "session.auth_session.userid", username);
      objectPath.set(req, "session.auth_session.first_factor", true);
      logger.info("1st factor: LDAP binding successful");
      logger.debug("1st factor: Retrieve email from LDAP");
      return BluebirdPromise.join(ldap.get_emails(username), ldap.get_groups(username));
    })
    .then(function (data: [string[], string[]]) {
      const emails: string[] = data[0];
      const groups: string[] = data[1];

      if (!emails && emails.length <= 0) throw new Error("No email found");
      logger.debug("1st factor: Retrieved email are %s", emails);
      objectPath.set(req, "session.auth_session.email", emails[0]);
      objectPath.set(req, "session.auth_session.groups", groups);

      regulator.mark(username, true);
      res.status(204);
      res.send();
    })
    .catch(exceptions.LdapSeachError, function (err: Error) {
      logger.error("1st factor: Unable to retrieve email from LDAP", err);
      res.status(500);
      res.send();
    })
    .catch(exceptions.LdapBindError, function (err: Error) {
      logger.error("1st factor: LDAP binding failed");
      logger.debug("1st factor: LDAP binding failed due to ", err);
      regulator.mark(username, false);
      res.status(401);
      res.send("Bad credentials");
    })
    .catch(exceptions.AuthenticationRegulationError, function (err: Error) {
      logger.error("1st factor: the regulator rejected the authentication of user %s", username);
      logger.debug("1st factor: authentication rejected due to  %s", err);
      res.status(403);
      res.send("Access has been restricted for a few minutes...");
    })
    .catch(exceptions.DomainAccessDenied, (err: Error) => {
      logger.error("1st factor: ", err);
      res.status(401);
      res.send("Access denied...");
    })
    .catch(function (err: Error) {
      console.log(err.stack);
      logger.error("1st factor: Unhandled error %s", err);
      res.status(500);
      res.send("Internal error");
    });
};
