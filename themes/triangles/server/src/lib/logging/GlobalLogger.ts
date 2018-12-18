import { IGlobalLogger } from "./IGlobalLogger";
import Util = require("util");
import Express = require("express");
import Winston = require("winston");

declare module "express" {
  interface Request {
    id: string;
  }
}

export class GlobalLogger implements IGlobalLogger {
  private winston: typeof Winston;
  constructor(winston: typeof Winston) {
    this.winston = winston;
  }

  private buildMessage(message: string, ...args: any[]): string {
    return Util.format("date='%s' message='%s'", new Date(),
      Util.format(message, ...args));
  }

  info(message: string, ...args: any[]): void {
    this.winston.info(this.buildMessage(message, ...args));
  }

  debug(message: string, ...args: any[]): void {
    this.winston.debug(this.buildMessage(message, ...args));
  }

  error(message: string, ...args: any[]): void {
    this.winston.debug(this.buildMessage(message, ...args));
  }
}