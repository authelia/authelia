import { IRequestLogger } from "../../src/lib/logging/IRequestLogger";
import Sinon = require("sinon");

export class RequestLoggerStub implements IRequestLogger {
  infoStub: Sinon.SinonStub;
  debugStub: Sinon.SinonStub;
  errorStub: Sinon.SinonStub;

  constructor() {
    this.infoStub = Sinon.stub();
    this.debugStub = Sinon.stub();
    this.errorStub = Sinon.stub();
  }

  info(req: Express.Request, message: string, ...args: any[]): void {
    return this.infoStub(req, message, ...args);
  }

  debug(req: Express.Request, message: string, ...args: any[]): void {
    return this.debugStub(req, message, ...args);
  }

  error(req: Express.Request, message: string, ...args: any[]): void {
    return this.errorStub(req, message, ...args);
  }
}