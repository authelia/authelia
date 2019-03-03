
import BluebirdPromise = require("bluebird");
import express = require("express");
import Exceptions = require("./Exceptions");
import { AuthenticationSessionHandler } from "./AuthenticationSessionHandler";
import { IRequestLogger } from "./logging/IRequestLogger";
import { Level } from "./authentication/Level";

export function validate(req: express.Request, logger: IRequestLogger): BluebirdPromise<void> {
  return new BluebirdPromise(function (resolve, reject) {
    const authSession = AuthenticationSessionHandler.get(req, logger);

    if (!authSession.userid || authSession.authentication_level < Level.ONE_FACTOR)
      return reject(new Exceptions.FirstFactorValidationError(
        "First factor has not been validated yet."));

    resolve();
  });
}