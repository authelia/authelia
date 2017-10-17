
import express = require("express");
import objectPath = require("object-path");
import winston = require("winston");
import Endpoints = require("../../../../../shared/api");
import { ServerVariables } from "../../ServerVariables";
import AuthenticationSession = require("../../AuthenticationSession");
import BluebirdPromise = require("bluebird");
import ErrorReplies = require("../../ErrorReplies");
import UserMessages = require("../../../../../shared/UserMessages");
import { RedirectionMessage } from "../../../../../shared/RedirectionMessage";
import Constants = require("../../../../../shared/constants");

export default function (vars: ServerVariables) {
  return function (req: express.Request, res: express.Response): BluebirdPromise<void> {
    return AuthenticationSession.get(req, vars.logger)
      .then(function (authSession) {
        let redirectUrl: string;
        if (vars.config.default_redirection_url) {
          redirectUrl = vars.config.default_redirection_url;
        }
        vars.logger.debug(req, "Request redirection to \"%s\".", redirectUrl);
        res.json({
          redirect: redirectUrl
        } as RedirectionMessage);
        return BluebirdPromise.resolve();
      })
      .catch(ErrorReplies.replyWithError200(req, res, vars.logger,
        UserMessages.OPERATION_FAILED));
  };
}