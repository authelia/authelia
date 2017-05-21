
import objectPath = require("object-path");
import express = require("express");

type ExpressRequest = (req: express.Request, res: express.Response, next?: express.NextFunction) => void;

export = function(callback: ExpressRequest): ExpressRequest {
  return function (req: express.Request, res: express.Response, next: express.NextFunction) {
    const auth_session = req.session.auth_session;
    const first_factor = objectPath.has(req, "session.auth_session.first_factor")
      && req.session.auth_session.first_factor;
    if (!first_factor) {
      res.status(403);
      res.send();
      return;
    }
    callback(req, res, next);
  };
};
