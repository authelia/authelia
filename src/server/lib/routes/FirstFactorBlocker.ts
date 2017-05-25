
import express = require("express");
import BluebirdPromise = require("bluebird");
import FirstFactorValidator = require("../FirstFactorValidator");
import Exceptions = require("../Exceptions");
import ErrorReplies = require("../ErrorReplies");
import objectPath = require("object-path");
import ServerVariables = require("../ServerVariables");
import AuthenticationSession = require("../AuthenticationSession");

type Handler = (req: express.Request, res: express.Response) => BluebirdPromise<void>;

export default function (callback: Handler): Handler {
    return function (req: express.Request, res: express.Response): BluebirdPromise<void> {
        const logger = ServerVariables.getLogger(req.app);

        const authSession = AuthenticationSession.get(req);
        logger.debug("AuthSession is %s", JSON.stringify(authSession));
        return FirstFactorValidator.validate(req)
            .then(function () {
                return callback(req, res);
            })
            .catch(Exceptions.FirstFactorValidationError, ErrorReplies.replyWithError401(res, logger));
    };
}