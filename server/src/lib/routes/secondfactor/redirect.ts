
import express = require("express");
import objectPath = require("object-path");
import winston = require("winston");
import Endpoints = require("../../../../../shared/api");
import { ServerVariablesHandler } from "../../ServerVariablesHandler";
import AuthenticationSession = require("../../AuthenticationSession");
import BluebirdPromise = require("bluebird");
import ErrorReplies = require("../../ErrorReplies");
import UserMessages = require("../../../../../shared/UserMessages");

export default function (req: express.Request, res: express.Response): BluebirdPromise<void> {
    const logger = ServerVariablesHandler.getLogger(req.app);
    return AuthenticationSession.get(req)
        .then(function (authSession: AuthenticationSession.AuthenticationSession) {
            const redirectUrl = req.query.redirect || Endpoints.FIRST_FACTOR_GET;
            res.json({
                redirection_url: redirectUrl
            });
            return BluebirdPromise.resolve();
        })
        .catch(ErrorReplies.replyWithError200(req, res, logger,
            UserMessages.OPERATION_FAILED));
}