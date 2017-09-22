
import Express = require("express");
import Endpoints = require("../../../endpoints");
import FirstFactorBlocker = require("../FirstFactorBlocker");
import BluebirdPromise = require("bluebird");
import { ServerVariablesHandler } from "../../ServerVariablesHandler";

const TEMPLATE_NAME = "secondfactor";

export default FirstFactorBlocker.default(handler);

function handler(req: Express.Request, res: Express.Response): BluebirdPromise<void> {
    const logger = ServerVariablesHandler.getLogger(req.app);
    logger.debug("secondfactor request is coming from %s", req.originalUrl);
    res.render(TEMPLATE_NAME, {
        totp_identity_start_endpoint: Endpoints.SECOND_FACTOR_TOTP_IDENTITY_START_GET,
        u2f_identity_start_endpoint: Endpoints.SECOND_FACTOR_U2F_IDENTITY_START_GET
    });
    return BluebirdPromise.resolve();
}