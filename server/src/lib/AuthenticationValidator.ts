
import BluebirdPromise = require("bluebird");
import express = require("express");
import objectPath = require("object-path");
import FirstFactorValidator = require("./FirstFactorValidator");
import AuthenticationSessionHandler = require("./AuthenticationSession");
import { IRequestLogger } from "./logging/IRequestLogger";

export function validate(req: express.Request, logger: IRequestLogger): BluebirdPromise<void> {
    return FirstFactorValidator.validate(req, logger)
        .then(function () {
            return AuthenticationSessionHandler.get(req, logger);
        })
        .then(function (authSession) {
            if (!authSession.second_factor)
                return BluebirdPromise.reject("No second factor variable.");
            return BluebirdPromise.resolve();
        });
}