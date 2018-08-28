import { IRequestLogger } from "./IRequestLogger";
import Util = require("util");
import Express = require("express");
import Winston = require("winston");

declare module "express" {
  interface Request {
    id: string;
  }
}

export class RequestLogger implements IRequestLogger {
  private winston: typeof Winston;

  constructor(winston: typeof Winston) {
    this.winston = winston;
  }

  private formatHeader(req: Express.Request) {
    const clientIP = req.ip; // The IP of the original client going through the proxy chain.
    return Util.format("date='%s' method='%s', path='%s' requestId='%s' sessionId='%s' ip='%s'",
      new Date(), req.method, req.path, req.id, req.sessionID, clientIP);
  }

  private formatBody(message: string) {
    return Util.format("message='%s'", message);
  }

  private formatMessage(req: Express.Request, message: string) {
    return Util.format("%s %s", this.formatHeader(req),
      this.formatBody(message));
  }

  info(req: Express.Request, message: string, ...args: any[]): void {
    this.winston.info(this.formatMessage(req, message), ...args);
  }

  debug(req: Express.Request, message: string, ...args: any[]): void {
    this.winston.debug(this.formatMessage(req, message), ...args);
  }

  error(req: Express.Request, message: string, ...args: any[]): void {
    this.winston.error(this.formatMessage(req, message), ...args);
  }
}