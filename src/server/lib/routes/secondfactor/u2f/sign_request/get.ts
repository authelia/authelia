
import objectPath = require("object-path");
import U2f = require("u2f");
import u2f_common = require("../../../secondfactor/u2f/U2FCommon");
import BluebirdPromise = require("bluebird");
import express = require("express");
import UserDataStore, { U2FRegistrationDocument } from "../../../../UserDataStore";
import { Winston } from "../../../../../../types/Dependencies";
import exceptions = require("../../../../Exceptions");
import { SignMessage } from "./SignMessage";
import FirstFactorBlocker from "../../../FirstFactorBlocker";
import ErrorReplies = require("../../../../ErrorReplies");
import ServerVariables = require("../../../../ServerVariables");
import AuthenticationSession = require("../../../../AuthenticationSession");

export default FirstFactorBlocker(handler);


export function handler(req: express.Request, res: express.Response): BluebirdPromise<void> {
    const logger = ServerVariables.getLogger(req.app);
    const userDataStore = ServerVariables.getUserDataStore(req.app);
    const authSession = AuthenticationSession.get(req);

    const userid = authSession.userid;
    const appid = u2f_common.extract_app_id(req);
    return userDataStore.get_u2f_meta(userid, appid)
        .then(function (doc: U2FRegistrationDocument): BluebirdPromise<SignMessage> {
            if (!doc)
                return BluebirdPromise.reject(new exceptions.AccessDeniedError("No U2F registration found"));

            const u2f = ServerVariables.getU2F(req.app);
            const appId: string = u2f_common.extract_app_id(req);
            logger.info("U2F sign_request: Start authentication to app %s", appId);
            logger.debug("U2F sign_request: appId=%s, keyHandle=%s", appId, JSON.stringify(doc.keyHandle));

            const request = u2f.request(appId, doc.keyHandle);
            const authenticationMessage: SignMessage = {
                request: request,
                keyHandle: doc.keyHandle
            };
            return BluebirdPromise.resolve(authenticationMessage);
        })
        .then(function (authenticationMessage: SignMessage) {
            logger.info("U2F sign_request: Store authentication request and reply");
            logger.debug("U2F sign_request: authenticationRequest=%s", authenticationMessage);
            authSession.sign_request = authenticationMessage.request;
            res.json(authenticationMessage);
            return BluebirdPromise.resolve();
        })
        .catch(exceptions.AccessDeniedError, ErrorReplies.replyWithError401(res, logger))
        .catch(ErrorReplies.replyWithError500(res, logger));
}

