
import express = require("express");
import * as URLParse from "url-parse";
import * as ObjectPath from "object-path";
import { ServerVariables } from "../../ServerVariables";
import BluebirdPromise = require("bluebird");
import ErrorReplies = require("../../ErrorReplies");
import UserMessages = require("../../../../../shared/UserMessages");
import IsRedirectionSafe from "../../../lib/utils/IsRedirectionSafe";
import { AuthenticationSessionHandler } from "../../../lib/AuthenticationSessionHandler";


export default function (vars: ServerVariables) {
  return function (req: express.Request, res: express.Response)
    : BluebirdPromise<void> {

    return new BluebirdPromise<void>(function (resolve, reject) {
      let redirectUrl: string = ObjectPath.get<Express.Request, string>(
        req, "headers.x-target-url", undefined);

      if (!redirectUrl && vars.config.default_redirection_url) {
        redirectUrl = vars.config.default_redirection_url;
      }

      const authSession = AuthenticationSessionHandler.get(req, vars.logger);
      if ((redirectUrl && !IsRedirectionSafe(vars, authSession, new URLParse(redirectUrl)))
          || !redirectUrl) {
        res.status(204);
        res.send();
        return resolve();
      }

      res.json({redirect: redirectUrl});
      return resolve();
    })
      .catch(ErrorReplies.replyWithError200(req, res, vars.logger,
        UserMessages.OPERATION_FAILED));
  };
}
