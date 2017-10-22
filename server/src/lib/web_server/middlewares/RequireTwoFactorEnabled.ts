import Express = require("express");
import BluebirdPromise = require("bluebird");
import ErrorReplies = require("../../ErrorReplies");
import { IRequestLogger } from "../../logging/IRequestLogger";
import { MethodCalculator } from "../../authentication/MethodCalculator";
import { AuthenticationMethodsConfiguration } from
  "../../configuration/Configuration";

export class RequireTwoFactorEnabled {
  static middleware(logger: IRequestLogger,
    configuration: AuthenticationMethodsConfiguration) {

    return function (req: Express.Request, res: Express.Response,
      next: Express.NextFunction): void {

      const isSingleFactorMode = MethodCalculator.isSingleFactorOnlyMode(
        configuration);

      if (isSingleFactorMode) {
        ErrorReplies.replyWithError401(req, res, logger)(new Error(
          "Restricted access because server is in single factor mode."));
        return;
      }
      next();
    };
  }
}