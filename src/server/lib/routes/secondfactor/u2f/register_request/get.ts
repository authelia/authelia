
import { UserDataStore } from "../../../../storage/UserDataStore";

import objectPath = require("object-path");
import u2f_common = require("../U2FCommon");
import BluebirdPromise = require("bluebird");
import express = require("express");
import U2f = require("u2f");
import FirstFactorBlocker from "../../../FirstFactorBlocker";
import ErrorReplies = require("../../../../ErrorReplies");
import { ServerVariablesHandler } from "../../../../ServerVariablesHandler";
import AuthenticationSession = require("../../../../AuthenticationSession");

export default FirstFactorBlocker(handler);

function handler(req: express.Request, res: express.Response): BluebirdPromise<void> {
    const logger = ServerVariablesHandler.getLogger(req.app);
    const authSession = AuthenticationSession.get(req);

    if (!authSession.identity_check
        || authSession.identity_check.challenge != "u2f-register") {
        res.status(403);
        res.send();
        return;
    }

    const u2f = ServerVariablesHandler.getU2F(req.app);
    const appid: string = u2f_common.extract_app_id(req);

    logger.debug("U2F register_request: headers=%s", JSON.stringify(req.headers));
    logger.info("U2F register_request: Starting registration for appId %s", appid);

    return BluebirdPromise.resolve(u2f.request(appid))
        .then(function (registrationRequest: U2f.Request) {
            logger.debug("U2F register_request: registrationRequest = %s", JSON.stringify(registrationRequest));
            authSession.register_request = registrationRequest;
            res.json(registrationRequest);
            return BluebirdPromise.resolve();
        })
        .catch(ErrorReplies.replyWithError500(res, logger));
}