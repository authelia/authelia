import express = require("express");
import BluebirdPromise = require("bluebird");
import { IRequestLogger } from "./logging/IRequestLogger";

function replyWithError(req: express.Request, res: express.Response,
  code: number, logger: IRequestLogger): (err: Error) => void {
  return function (err: Error): void {
    logger.error(req, "Reply with error %d: %s", code, err.message);
    logger.debug(req, "%s", err.stack);
    res.status(code);
    res.send();
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

export function replyWithError500(req: express.Request,
  res: express.Response, logger: IRequestLogger) {
  return replyWithError(req, res, 500, logger);
}