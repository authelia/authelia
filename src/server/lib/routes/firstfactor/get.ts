
import express = require("express");
import objectPath = require("object-path");
import winston = require("winston");
import Endpoints = require("../../../endpoints");
import AuthenticationValidator = require("../../AuthenticationValidator");
import ServerVariables = require("../../ServerVariables");

export default function (req: express.Request, res: express.Response) {
    const logger = ServerVariables.getLogger(req.app);

    logger.debug("First factor: headers are %s", JSON.stringify(req.headers));

    AuthenticationValidator.validate(req)
        .then(function () {
            res.render("already-logged-in", { logout_endpoint: Endpoints.LOGOUT_GET });
        }, function () {
            res.render("firstfactor", {
                first_factor_post_endpoint: Endpoints.FIRST_FACTOR_POST,
                reset_password_request_endpoint: Endpoints.RESET_PASSWORD_REQUEST_GET
            });
        });
}