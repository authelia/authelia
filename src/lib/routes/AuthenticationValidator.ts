
import objectPath = require("object-path");
import BluebirdPromise = require("bluebird");
import express = require("express");

function verify_filter(req: express.Request, res: express.Response) {
  const logger = req.app.get("logger");

  if (!objectPath.has(req, "session.auth_session"))
    return BluebirdPromise.reject("No auth_session variable");

  if (!objectPath.has(req, "session.auth_session.first_factor"))
    return BluebirdPromise.reject("No first factor variable");

  if (!objectPath.has(req, "session.auth_session.second_factor"))
    return BluebirdPromise.reject("No second factor variable");

  if (!objectPath.has(req, "session.auth_session.userid"))
    return BluebirdPromise.reject("No userid variable");

  const host = objectPath.get<express.Request, string>(req, "headers.host");
  const domain = host.split(":")[0];

  if (!req.session.auth_session.first_factor ||
    !req.session.auth_session.second_factor)
    return BluebirdPromise.reject("First or second factor not validated");

  return BluebirdPromise.resolve();
}

export = function(req: express.Request, res: express.Response) {
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

