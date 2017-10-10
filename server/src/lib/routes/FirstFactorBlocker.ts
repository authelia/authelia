
import express = require("express");
import BluebirdPromise = require("bluebird");
import FirstFactorValidator = require("../FirstFactorValidator");
import Exceptions = require("../Exceptions");
import ErrorReplies = require("../ErrorReplies");
import objectPath = require("object-path");
import { ServerVariablesHandler } from "../ServerVariablesHandler";
import AuthenticationSession = require("../AuthenticationSession");
import UserMessages = require("../../../../shared/UserMessages");

type Handler = (req: express.Request, res: express.Response) => BluebirdPromise<void>;

export default function (callback: Handler): Handler {
  return function (req: express.Request, res: express.Response): BluebirdPromise<void> {
    const logger = ServerVariablesHandler.getLogger(req.app);

    return AuthenticationSession.get(req)
      .then(function (authSession) {
        return FirstFactorValidator.validate(req);
      })
      .then(function () {
        return callback(req, res);
      })
      .catch(ErrorReplies.replyWithError401(req, res, logger));
  };
}