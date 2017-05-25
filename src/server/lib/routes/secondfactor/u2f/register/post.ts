
import UserDataStore from "../../../../UserDataStore";

import objectPath = require("object-path");
import u2f_common = require("../U2FCommon");
import BluebirdPromise = require("bluebird");
import express = require("express");
import U2f = require("u2f");
import FirstFactorBlocker from "../../../FirstFactorBlocker";
import redirect from "../../redirect";
import ErrorReplies = require("../../../../ErrorReplies");
import ServerVariables = require("../../../../ServerVariables");
import AuthenticationSession = require("../../../../AuthenticationSession");


export default FirstFactorBlocker(handler);


function handler(req: express.Request, res: express.Response): BluebirdPromise<void> {
    const authSession = AuthenticationSession.get(req);
    const registrationRequest = authSession.register_request;

    if (!registrationRequest) {
        res.status(403);
        res.send();
        return BluebirdPromise.reject(new Error("No registration request"));
    }

    if (!authSession.identity_check
        || authSession.identity_check.challenge != "u2f-register") {
        res.status(403);
        res.send();
        return BluebirdPromise.reject(new Error("Bad challenge for registration request"));
    }


    const userDataStore = ServerVariables.getUserDataStore(req.app);
    const u2f = ServerVariables.getU2F(req.app);
    const userid = authSession.userid;
    const appid = u2f_common.extract_app_id(req);
    const logger = ServerVariables.getLogger(req.app);

    const registrationResponse: U2f.RegistrationData = req.body;

    logger.info("U2F register: Finishing registration");
    logger.debug("U2F register: registrationRequest = %s", JSON.stringify(registrationRequest));
    logger.debug("U2F register: registrationResponse = %s", JSON.stringify(registrationResponse));

    BluebirdPromise.resolve(u2f.checkRegistration(registrationRequest, registrationResponse))
        .then(function (u2fResult: U2f.RegistrationResult | U2f.Error): BluebirdPromise<void> {
            if (objectPath.has(u2fResult, "errorCode"))
                return BluebirdPromise.reject(new Error("Error while registering."));

            const registrationResult: U2f.RegistrationResult = u2fResult as U2f.RegistrationResult;
            logger.info("U2F register: Store regisutration and reply");
            logger.debug("U2F register: registration = %s", JSON.stringify(registrationResult));
            return userDataStore.set_u2f_meta(userid, appid, registrationResult.keyHandle, registrationResult.publicKey);
        })
        .then(function () {
            authSession.identity_check = undefined;
            redirect(req, res);
            return BluebirdPromise.resolve();
        })
        .catch(ErrorReplies.replyWithError500(res, logger));
}
