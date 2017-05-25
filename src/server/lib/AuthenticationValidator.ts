
import BluebirdPromise = require("bluebird");
import express = require("express");
import objectPath = require("object-path");

import FirstFactorValidator = require("./FirstFactorValidator");
import AuthenticationSession = require("./AuthenticationSession");

export function validate(req: express.Request): BluebirdPromise<void> {
    return FirstFactorValidator.validate(req)
        .then(function () {
            const authSession = AuthenticationSession.get(req);
            if (!authSession.second_factor)
                return BluebirdPromise.reject("No second factor variable");

            return BluebirdPromise.resolve();
        });
}