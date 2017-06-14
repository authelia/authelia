
import exceptions = require("../../Exceptions");
import objectPath = require("object-path");
import BluebirdPromise = require("bluebird");
import express = require("express");
import { AccessController } from "../../access_control/AccessController";
import { AuthenticationRegulator } from "../../AuthenticationRegulator";
import { LdapClient } from "../../LdapClient";
import Endpoint = require("../../../endpoints");
import ErrorReplies = require("../../ErrorReplies");
import ServerVariables = require("../../ServerVariables");
import AuthenticationSession = require("../../AuthenticationSession");

export default function (req: express.Request, res: express.Response): BluebirdPromise<void> {
    const username: string = req.body.username;
    const password: string = req.body.password;

    const logger = ServerVariables.getLogger(req.app);
    const ldap = ServerVariables.getLdapClient(req.app);
    const config = ServerVariables.getConfiguration(req.app);

    if (!username || !password) {
        const err = new Error("No username or password");
        ErrorReplies.replyWithError401(res, logger)(err);
        return BluebirdPromise.reject(err);
    }

    const regulator = ServerVariables.getAuthenticationRegulator(req.app);
    const accessController = ServerVariables.getAccessController(req.app);
    const authSession = AuthenticationSession.get(req);

    logger.info("1st factor: Starting authentication of user \"%s\"", username);
    logger.debug("1st factor: Start bind operation against LDAP");
    logger.debug("1st factor: username=%s", username);

    return regulator.regulate(username)
        .then(function () {
            return ldap.bind(username, password);
        })
        .then(function () {
            authSession.userid = username;
            authSession.first_factor = true;
            logger.info("1st factor: LDAP binding successful");
            logger.debug("1st factor: Retrieve email from LDAP");
            return BluebirdPromise.join(ldap.get_emails(username), ldap.get_groups(username));
        })
        .then(function (data: [string[], string[]]) {
            const emails: string[] = data[0];
            const groups: string[] = data[1];

            if (!emails && emails.length <= 0) throw new Error("No email found");
            logger.debug("1st factor: Retrieved email are %s", emails);
            authSession.email = emails[0];
            authSession.groups = groups;

            regulator.mark(username, true);
            logger.debug("1st factor: Redirect to  %s", Endpoint.SECOND_FACTOR_GET);
            res.redirect(Endpoint.SECOND_FACTOR_GET);
            return BluebirdPromise.resolve();
        })
        .catch(exceptions.LdapSearchError, ErrorReplies.replyWithError500(res, logger))
        .catch(exceptions.LdapBindError, function (err: Error) {
            regulator.mark(username, false);
            ErrorReplies.replyWithError401(res, logger)(err);
        })
        .catch(exceptions.AuthenticationRegulationError, ErrorReplies.replyWithError403(res, logger))
        .catch(exceptions.DomainAccessDenied, ErrorReplies.replyWithError401(res, logger))
        .catch(ErrorReplies.replyWithError500(res, logger));
}
