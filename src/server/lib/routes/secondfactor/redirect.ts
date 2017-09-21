
import express = require("express");
import objectPath = require("object-path");
import winston = require("winston");
import Endpoints = require("../../../endpoints");
import { ServerVariablesHandler } from "../../ServerVariablesHandler";
import AuthenticationSession = require("../../AuthenticationSession");
import BluebirdPromise = require("bluebird");

export default function (req: express.Request, res: express.Response): BluebirdPromise<void> {
    return AuthenticationSession.get(req)
        .then(function (authSession: AuthenticationSession.AuthenticationSession) {
            const redirectUrl = req.query.redirect || Endpoints.FIRST_FACTOR_GET;
            res.json({
                redirection_url: redirectUrl
            });
            return BluebirdPromise.resolve();
        });
}