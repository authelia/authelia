import Sinon = require("sinon");
import { GlobalLogger } from "./GlobalLogger";
import Winston = require("winston");
import Express = require("express");
import { IGlobalLogger } from "./IGlobalLogger";

export class GlobalLoggerStub implements IGlobalLogger {
  infoStub: Sinon.SinonStub;
  debugStub: Sinon.SinonStub;
  errorStub: Sinon.SinonStub;
  private globalLogger: IGlobalLogger;

  constructor(enableLogging?: boolean) {
    this.infoStub = Sinon.stub();
    this.debugStub = Sinon.stub();
    this.errorStub = Sinon.stub();
    if (enableLogging)
      this.globalLogger = new GlobalLogger(Winston);
  }

  info(message: string, ...args: any[]): void {
    if (this.globalLogger)
      this.globalLogger.info(message, ...args);
    this.infoStub(message, ...args);
  }

  debug(message: string, ...args: any[]): void {
    if (this.globalLogger)
      this.globalLogger.info(message, ...args);
    this.debugStub(message, ...args);
  }

  error(message: string, ...args: any[]): void {
    if (this.globalLogger)
      this.globalLogger.info(message, ...args);
    this.errorStub(message, ...args);
  }
}