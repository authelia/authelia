import * as Express from "express";
import { IRequestLogger } from "./logging/IRequestLogger";

function replyWithError(req: Express.Request, res: Express.Response,
  code: number, logger: IRequestLogger, body?: Object): (err: Error) => void {
  return function (err: Error): void {
    logger.error(req, "Reply with error %d: %s", code, err.message);
    logger.debug(req, "%s", err.stack);
    res.status(code);
    res.send(body);
  };
}

export function redirectTo(redirectUrl: string, req: Express.Request,
  res: Express.Response, logger: IRequestLogger) {
  return function(err: Error) {
    logger.error(req, "Error: %s", err.message);
    logger.debug(req, "Redirecting to %s", redirectUrl);
    res.redirect(redirectUrl);
  };
}

export function replyWithError400(req: Express.Request,
  res: Express.Response, logger: IRequestLogger) {
  return replyWithError(req, res, 400, logger);
}

export function replyWithError401(req: Express.Request,
  res: Express.Response, logger: IRequestLogger) {
  return replyWithError(req, res, 401, logger);
}

export function replyWithError403(req: Express.Request,
  res: Express.Response, logger: IRequestLogger) {
  return replyWithError(req, res, 403, logger);
}

export function replyWithError200(req: Express.Request,
  res: Express.Response, logger: IRequestLogger, message: string) {
  return replyWithError(req, res, 200, logger, { error: message });
}