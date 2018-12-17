import Express = require("express");

export interface IRequestLogger {
  info(req: Express.Request, message: string, ...args: any[]): void;
  debug(req: Express.Request, message: string, ...args: any[]): void;
  error(req: Express.Request, message: string, ...args: any[]): void;
}