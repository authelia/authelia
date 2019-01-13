
import express = require("express");
import { ServerVariables } from "../../ServerVariables";
import BluebirdPromise = require("bluebird");
import ErrorReplies = require("../../ErrorReplies");
import UserMessages = require("../../../../../shared/UserMessages");
import { RedirectionMessage } from "../../../../../shared/RedirectionMessage";

export default function (vars: ServerVariables) {
  return function (req: express.Request, res: express.Response)
    : BluebirdPromise<void> {

    return new BluebirdPromise<void>(function (resolve, reject) {
      let redirectUrl: string = "/";
      if (vars.config.default_redirection_url) {
        redirectUrl = vars.config.default_redirection_url;
      }
      vars.logger.debug(req, "Request redirection to \"%s\".", redirectUrl);
      res.json({
        redirect: redirectUrl
      } as RedirectionMessage);
      return resolve();
    })
      .catch(ErrorReplies.replyWithError200(req, res, vars.logger,
        UserMessages.OPERATION_FAILED));
  };
}
