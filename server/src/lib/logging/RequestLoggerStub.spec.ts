import { IRequestLogger } from "./IRequestLogger";
import Sinon = require("sinon");
import { RequestLogger } from "./RequestLogger";
import Winston = require("winston");
import Express = require("express");

export class RequestLoggerStub implements IRequestLogger {
  infoStub: Sinon.SinonStub;
  debugStub: Sinon.SinonStub;
  errorStub: Sinon.SinonStub;
  private requestLogger: RequestLogger;

  constructor(enableLogging?: boolean) {
    this.infoStub = Sinon.stub();
    this.debugStub = Sinon.stub();
    this.errorStub = Sinon.stub();
    if (enableLogging)
      this.requestLogger = new RequestLogger(Winston);
  }

  info(req: Express.Request, message: string, ...args: any[]): void {
    if (this.requestLogger)
      this.requestLogger.info(req, message, ...args);
    this.infoStub(req, message, ...args);
  }

  debug(req: Express.Request, message: string, ...args: any[]): void {
    if (this.requestLogger)
      this.requestLogger.info(req, message, ...args);
    this.debugStub(req, message, ...args);
  }

  error(req: Express.Request, message: string, ...args: any[]): void {
    if (this.requestLogger)
      this.requestLogger.info(req, message, ...args);
    this.errorStub(req, message, ...args);
  }
}