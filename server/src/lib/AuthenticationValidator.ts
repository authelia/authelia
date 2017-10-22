
import BluebirdPromise = require("bluebird");
import express = require("express");
import objectPath = require("object-path");
import FirstFactorValidator = require("./FirstFactorValidator");
import { AuthenticationSessionHandler } from "./AuthenticationSessionHandler";
import { IRequestLogger } from "./logging/IRequestLogger";

export function validate(req: express.Request, logger: IRequestLogger): BluebirdPromise<void> {
  return FirstFactorValidator.validate(req, logger)
    .then(function () {
      const authSession = AuthenticationSessionHandler.get(req, logger);
      if (!authSession.second_factor)
        return BluebirdPromise.reject("No second factor variable.");
      return BluebirdPromise.resolve();
    });
}