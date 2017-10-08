
import Express = require("express");
import Endpoints = require("../../../../../shared/api");
import FirstFactorBlocker = require("../FirstFactorBlocker");
import BluebirdPromise = require("bluebird");
import { ServerVariablesHandler } from "../../ServerVariablesHandler";
import AuthenticationSession = require("../../AuthenticationSession");

const TEMPLATE_NAME = "secondfactor";

export default FirstFactorBlocker.default(handler);

function handler(req: Express.Request, res: Express.Response): BluebirdPromise<void> {
    return AuthenticationSession.get(req)
        .then(function (authSession) {
            if (authSession.first_factor && authSession.second_factor) {
                res.redirect(Endpoints.LOGGED_IN);
                return BluebirdPromise.resolve();
            }

            res.render(TEMPLATE_NAME, {
                username: authSession.userid,
                totp_identity_start_endpoint: Endpoints.SECOND_FACTOR_TOTP_IDENTITY_START_GET,
                u2f_identity_start_endpoint: Endpoints.SECOND_FACTOR_U2F_IDENTITY_START_GET
            });
            return BluebirdPromise.resolve();
        });
}