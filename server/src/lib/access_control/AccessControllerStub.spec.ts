import Sinon = require("sinon");
import { IAccessController } from "./IAccessController";

export class AccessControllerStub implements IAccessController {
  isWhitelistedMock: Sinon.SinonStub;
  isAccessAllowedMock: Sinon.SinonStub;

  constructor() {
    this.isWhitelistedMock = Sinon.stub();
    this.isAccessAllowedMock = Sinon.stub();
  }

  isWhitelisted(domain: string, ip: string): boolean {
    return this.isWhitelistedMock(domain, ip)
  }

  isAccessAllowed(domain: string, resource: string, user: string, groups: string[]): boolean {
    return this.isAccessAllowedMock(domain, resource, user, groups);
  }
}
