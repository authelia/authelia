import express = require("express");
import BluebirdPromise = require("bluebird");
import { IRequestLogger } from "./logging/IRequestLogger";

function replyWithError(req: express.Request, res: express.Response,
  code: number, logger: IRequestLogger, body?: Object): (err: Error) => void {
  return function (err: Error): void {
    if (req.originalUrl.startsWith("/api/") || code == 200) {
      logger.error(req, "Reply with error %d: %s", code, err.message);
      logger.debug(req, "%s", err.stack);
      res.status(code);
      res.send(body);
    }
    else {
      logger.error(req, "Redirect to error %d: %s", code, err.message);
      logger.debug(req, "%s", err.stack);
      res.redirect("/error/" + code);
    }
  };
}

export function replyWithError400(req: express.Request,
  res: express.Response, logger: IRequestLogger) {
  return replyWithError(req, res, 400, logger);
}

export function replyWithError401(req: express.Request,
  res: express.Response, logger: IRequestLogger) {
  return replyWithError(req, res, 401, logger);
}

export function replyWithError403(req: express.Request,
  res: express.Response, logger: IRequestLogger) {
  return replyWithError(req, res, 403, logger);
}

export function replyWithError200(req: express.Request,
  res: express.Response, logger: IRequestLogger, message: string) {
  return replyWithError(req, res, 200, logger, { error: message });
}