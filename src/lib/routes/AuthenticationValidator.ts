
import objectPath = require("object-path");
import BluebirdPromise = require("bluebird");
import express = require("express");
import AccessController from "../access_control/AccessController";
import exceptions = require("../Exceptions");

function verify_filter(req: express.Request, res: express.Response) {
  const logger = req.app.get("logger");
  const accessController: AccessController = req.app.get("access controller");

  if (!objectPath.has(req, "session.auth_session"))
    return BluebirdPromise.reject("No auth_session variable");

  if (!objectPath.has(req, "session.auth_session.first_factor"))
    return BluebirdPromise.reject("No first factor variable");

  if (!objectPath.has(req, "session.auth_session.second_factor"))
    return BluebirdPromise.reject("No second factor variable");

  if (!objectPath.has(req, "session.auth_session.userid"))
    return BluebirdPromise.reject("No userid variable");

  const username = objectPath.get<express.Request, string>(req, "session.auth_session.userid");
  const groups = objectPath.get<express.Request, string[]>(req, "session.auth_session.groups");

  const host = objectPath.get<express.Request, string>(req, "headers.host");
  const domain = host.split(":")[0];

  const isAllowed = accessController.isDomainAllowedForUser(domain, username, groups);
  if (!isAllowed) return BluebirdPromise.reject(
    new exceptions.DomainAccessDenied("User '" + username + "' does not have access to " + domain));

  if (!req.session.auth_session.first_factor ||
    !req.session.auth_session.second_factor)
    return BluebirdPromise.reject(new exceptions.AccessDeniedError("First or second factor not validated"));

  return BluebirdPromise.resolve();
}

export = function (req: express.Request, res: express.Response) {
  verify_filter(req, res)
    .then(function () {
      res.status(204);
      res.send();
    })
    .catch(function (err) {
      req.app.get("logger").error(err);
      res.status(401);
      res.send();
    });
};

