
import BluebirdPromise = require("bluebird");
import express = require("express");
import objectPath = require("object-path");
import Exceptions = require("./Exceptions");
import AuthenticationSession = require("./AuthenticationSession");

export function validate(req: express.Request): BluebirdPromise<void> {
    const authSession = AuthenticationSession.get(req);
    if (!authSession.userid || !authSession.first_factor)
        return BluebirdPromise.reject(new Exceptions.FirstFactorValidationError("First factor has not been validated yet."));

    return BluebirdPromise.resolve();
}