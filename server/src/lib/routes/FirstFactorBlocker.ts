
import express = require("express");
import BluebirdPromise = require("bluebird");
import FirstFactorValidator = require("../FirstFactorValidator");
import Exceptions = require("../Exceptions");
import ErrorReplies = require("../ErrorReplies");
import objectPath = require("object-path");
import AuthenticationSession = require("../AuthenticationSession");
import UserMessages = require("../../../../shared/UserMessages");
import { IRequestLogger } from "../logging/IRequestLogger";

type Handler = (req: express.Request, res: express.Response) => BluebirdPromise<void>;

export default function (callback: Handler, logger: IRequestLogger): Handler {
  return function (req: express.Request, res: express.Response): BluebirdPromise<void> {
    return AuthenticationSession.get(req, logger)
      .then(function (authSession) {
        return FirstFactorValidator.validate(req, logger);
      })
      .then(function () {
        return callback(req, res);
      })
      .catch(ErrorReplies.replyWithError401(req, res, logger));
  };
}