import Express = require("express");
import BluebirdPromise = require("bluebird");
import Util = require("util");

import { ServerVariables } from "../../ServerVariables";
import Exceptions = require("../../Exceptions");

export default function (req: Express.Request, vars: ServerVariables,
  domain: string, path: string, username: string, groups: string[]) {

  return new BluebirdPromise(function (resolve, reject) {
    const isAllowed = vars.accessController
      .isAccessAllowed(domain, path, username, groups);

    if (!isAllowed) {
      reject(new Exceptions.DomainAccessDenied(Util.format(
        "User '%s' does not have access to '%s'", username, domain)));
      return;
    }
    resolve();
  });
}