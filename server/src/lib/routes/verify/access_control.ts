import Express = require("express");
import BluebirdPromise = require("bluebird");
import Util = require("util");
import Exceptions = require("../../Exceptions");
import { ServerVariables } from "../../ServerVariables";
import { MethodCalculator } from "../../authentication/MethodCalculator";
import { WhitelistValue } from "../../authentication/whitelist/WhitelistHandler";

export default function (req: Express.Request, vars: ServerVariables,
  domain: string, path: string, username: string, groups: string[], whitelisted: WhitelistValue) {

  return new BluebirdPromise(function (resolve, reject) {
    const authenticationMethod =
      MethodCalculator.compute(vars.config.authentication_methods, domain);

    const isSecondFactorRequired = authenticationMethod === "two_factor";

    const isAllowed = vars.accessController
      .isAccessAllowed(domain, path, username, groups, whitelisted, isSecondFactorRequired);

    if (!isAllowed) {
      if (whitelisted) {
        if (isSecondFactorRequired) {
          return reject(new Exceptions.AccessDeniedError(Util.format(
            "Whitelisted user \"%s\" must perform second factor authentication for \"%s\"", username, domain)));
        }
        return reject(new Exceptions.AccessDeniedError(Util.format(
          "Whitelisted user \"%s\" must perform authentication on \"%s\"", username, domain)));
      }

      return reject(new Exceptions.DomainAccessDenied(Util.format(
        "User '%s' does not have access to '%s'", username, domain)));
    }
    resolve();
  });
}