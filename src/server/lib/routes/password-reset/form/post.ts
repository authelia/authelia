
import express = require("express");
import BluebirdPromise = require("bluebird");
import objectPath = require("object-path");
import exceptions = require("../../../Exceptions");
import ServerVariables = require("../../../ServerVariables");
import AuthenticationSession = require("../../../AuthenticationSession");
import ErrorReplies = require("../../../ErrorReplies");

import Constants = require("./../constants");

export default function (req: express.Request, res: express.Response): BluebirdPromise<void> {
    const logger = ServerVariables.getLogger(req.app);
    const ldap = ServerVariables.getLdapClient(req.app);
    const authSession = AuthenticationSession.get(req);

    const new_password = objectPath.get<express.Request, string>(req, "body.password");

    const userid = authSession.identity_check.userid;
    const challenge = authSession.identity_check.challenge;
    if (challenge != Constants.CHALLENGE) {
        res.status(403);
        res.send();
        return;
    }

    logger.info("POST reset-password: User %s wants to reset his/her password", userid);

    return ldap.update_password(userid, new_password)
        .then(function () {
            logger.info("POST reset-password: Password reset for user '%s'", userid);
            AuthenticationSession.reset(req);
            res.status(204);
            res.send();
            return BluebirdPromise.resolve();
        })
        .catch(ErrorReplies.replyWithError500(res, logger));
}
