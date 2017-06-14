import express = require("express");
import { Winston } from "winston";

function replyWithError(res: express.Response, code: number, logger: Winston) {
  return function (err: Error) {
    logger.error("Reply with error %d: %s", code, err);
    res.status(code);
    res.send();
  };
}

export function replyWithError400(res: express.Response, logger: Winston) {
  return replyWithError(res, 400, logger);
}

export function replyWithError401(res: express.Response, logger: Winston) {
  return replyWithError(res, 401, logger);
}

export function replyWithError403(res: express.Response, logger: Winston) {
  return replyWithError(res, 403, logger);
}

export function replyWithError500(res: express.Response, logger: Winston) {
  return replyWithError(res, 500, logger);
}